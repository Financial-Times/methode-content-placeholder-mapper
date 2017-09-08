package mapper

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

// Mapper is a generic interface for content placeholder mapper
type Mapper interface {
	HandlePlaceholderMessages(msg consumer.Message)
	StartMappingMessages(c consumer.MessageConsumer, p producer.MessageProducer)
	MapContentPlaceholder(mpc *model.MethodeContentPlaceholder) (*model.UppContentPlaceholder, *model.UppComplementaryContent, *utility.MappingError)
}

type defaultMapper struct {
	messageConsumer consumer.MessageConsumer
	messageProducer producer.MessageProducer
	cphValidator    CPHValidator
}

// New returns a new Mapper instance
func NewDefaultMapper() *defaultMapper {
	return &defaultMapper{cphValidator: NewDefaultCPHValidator()}
}

func (m *defaultMapper) HandlePlaceholderMessages(msg consumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if msg.Headers["Origin-System-Id"] != model.MethodeSystemID {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}

	placeholderMsg, placeholderUUID, complementaryMsg, complementaryUUID, mappingErr := m.mapMessage(msg)
	if mappingErr != nil {
		log.WithField("transaction_id", tid).WithField("uuid", mappingErr.ContentUUID).WithError(mappingErr).Warn("Error in mapping message")
		return
	}

	err := m.messageProducer.SendMessage(placeholderUUID, *placeholderMsg)
	if err != nil {
		log.WithField("transaction_id", tid).WithField("uuid", placeholderUUID).WithError(err).Warn("Error sending transformed content message to queue")
		return
	}

	err = m.messageProducer.SendMessage(complementaryUUID, *complementaryMsg)
	if err != nil {
		log.WithField("transaction_id", tid).WithField("uuid", complementaryUUID).WithError(err).Warn("Error sending transformed complementarycontent message to queue")
		return
	}

	log.WithField("transaction_id", tid).WithField("uuid", placeholderUUID).Info("Content mapped and sent to the queue")
}

func (m *defaultMapper) mapMessage(msg consumer.Message) (*producer.Message, string, *producer.Message, string, *utility.MappingError) {
	methodePlaceholder, err := model.NewMethodeContentPlaceholder([]byte(msg.Body), msg.Headers["X-Request-Id"], msg.Headers["Message-Timestamp"])
	if err != nil {
		return nil, "", nil, "", err
	}

	uppPlaceholder, uppComplementaryContent, err := m.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		return nil, "", nil, "", err
	}

	pubContentEventMsg, err := uppPlaceholder.ToPublicationEventMessage(uppPlaceholder)
	if err != nil {
		return nil, "", nil, "", err
	}

	pubComplementaryContentEventMsg, err := uppComplementaryContent.ToPublicationEventMessage(uppComplementaryContent)
	if err != nil {
		return nil, "", nil, "", err
	}

	return pubContentEventMsg, uppPlaceholder.UUID, pubComplementaryContentEventMsg, uppComplementaryContent.UUID, nil
}

func (m *defaultMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder) (*model.UppContentPlaceholder, *model.UppComplementaryContent, *utility.MappingError) {
	err := m.cphValidator.Validate(mcp)
	if err != nil {
		return nil, nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(mcp.UUID)
	}

	if mcp.IsInternalCPH() {
		if mcp.Attributes.IsDeleted {
			return nil, model.NewUppComplementaryContentDelete(mcp), nil
		}

		return nil, model.NewUppComplementaryContent(mcp, mcp.Attributes.LinkedArticleUUID), nil
	} else {
		if mcp.Attributes.IsDeleted {
			return model.NewUppContentPlaceholderDelete(mcp), model.NewUppComplementaryContentDelete(mcp), nil
		}

		uppContent, err := model.NewUppContentPlaceholder(mcp)
		if err != nil {
			return nil, nil, err
		}

		return uppContent, model.NewUppComplementaryContent(mcp, mcp.UUID), nil
	}
}

func (m *defaultMapper) StartMappingMessages(c consumer.MessageConsumer, p producer.MessageProducer) {
	m.messageConsumer = c
	m.messageProducer = p

	log.Infof("Starting queue consumer...")
	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		m.messageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	m.messageConsumer.Stop()
	consumerWaitGroup.Wait()
}
