package handler

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

type MessageHandler interface {
	HandleMessage(msg consumer.Message)
	StartHandlingMessages()
}

type CPHMessageHandler struct {
	MessageConsumer consumer.MessageConsumer
	messageProducer producer.MessageProducer
	nativeMapper    mapper.MessageToContentPlaceholderMapper
	cphMapper       mapper.CPHAggregateMapper
	messageCreator  message.MessageCreator
}

func NewCPHMessageHandler(c consumer.MessageConsumer,
	p producer.MessageProducer,
	mapper mapper.CPHAggregateMapper,
	nativeMapper mapper.MessageToContentPlaceholderMapper,
	messageCreator message.MessageCreator) *CPHMessageHandler {

	return &CPHMessageHandler{
		MessageConsumer : c,
		messageProducer : p,
		nativeMapper: nativeMapper,
		cphMapper:  mapper,
		messageCreator:   messageCreator,
	}
}

func (kqh *CPHMessageHandler) HandleMessage(msg consumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if msg.Headers["Origin-System-Id"] != model.MethodeSystemID {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}

	methodePlaceholder, err := kqh.nativeMapper.Map([]byte(msg.Body), msg.Headers["X-Request-Id"], msg.Headers["Message-Timestamp"])
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Error("Error creating methode model from queue message")
		return
	}

	transformedContents, err := kqh.cphMapper.MapContentPlaceholder(methodePlaceholder, tid)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Error("Error transforming content")
		return
	}

	for _, transformedContent := range transformedContents {
		eventMessage, err := kqh.messageCreator.ToPublicationEventMessage(transformedContent.GetUppCoreContent(), transformedContent)
		if err != nil {
			log.WithField("transaction_id", tid).WithField("uuid", transformedContent.GetUUID()).WithError(err).Warn("Error creating transformed content message to queue")
			return
		}

		rawErr := kqh.messageProducer.SendMessage("", *eventMessage)
		if rawErr != nil {
			log.WithField("transaction_id", tid).WithField("uuid", transformedContent.GetUUID()).WithError(err).Warn("Error sending transformed content message to queue")
			return
		}

		log.WithField("transaction_id", tid).WithField("uuid", transformedContent.GetUUID()).Info("Content mapped and sent to the queue")
	}
}

func (kqh *CPHMessageHandler) StartHandlingMessages() {
	log.Infof("Starting queue consumer...")
	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		kqh.MessageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	kqh.MessageConsumer.Stop()
	consumerWaitGroup.Wait()
}
