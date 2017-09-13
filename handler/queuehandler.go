package handler

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
)

type MessageHandler interface {
	HandleMessage(msg consumer.Message)
	StartHandlingMessages(c consumer.MessageConsumer, p producer.MessageProducer)
}

type CPHMessageHandler struct {
	messageConsumer   consumer.MessageConsumer
	messageProducer   producer.MessageProducer
	aggregateMapper   *mapper.AggregateCPHMapper
	cphMessageCreator *message.CPHMessageCreator
}

func NewCPHMessageHandler() *CPHMessageHandler {
	return &CPHMessageHandler{aggregateMapper: mapper.NewAggregateCPHMapper(), cphMessageCreator: message.NewDefaultCPHMessageCreator()}
}

func (kqh *CPHMessageHandler) HandleMessage(msg consumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if msg.Headers["Origin-System-Id"] != model.MethodeSystemID {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}

	methodePlaceholder, err := model.NewMethodeContentPlaceholder([]byte(msg.Body), msg.Headers["X-Request-Id"], msg.Headers["Message-Timestamp"])
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Error("Error creating methode model from queue message")
		return
	}

	transformedContents, err := kqh.aggregateMapper.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Error("Error transforming content")
		return
	}

	for _, transformedContent := range transformedContents {
		eventMessage, err := kqh.cphMessageCreator.ToPublicationEventMessage(transformedContent.GetUppCoreContent(), transformedContent)
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

func (kqh *CPHMessageHandler) StartHandlingMessages(c consumer.MessageConsumer, p producer.MessageProducer) {
	kqh.messageConsumer = c
	kqh.messageProducer = p

	log.Infof("Starting queue consumer...")
	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		kqh.messageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	kqh.messageConsumer.Stop()
	consumerWaitGroup.Wait()
}
