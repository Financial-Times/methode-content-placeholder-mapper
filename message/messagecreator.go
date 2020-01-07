package message

import (
	"encoding/json"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-TimesFinancial-Times/methode-content-placeholder-mapper/v2/model"
	"github.com/satori/go.uuid"
)

type MessageCreator interface {
	ToPublicationEventMessage(coreAttributes *model.UppCoreContent, payload interface{}) (*producer.Message, error)
	ToPublicationEvent(coreAttributes *model.UppCoreContent, payload interface{}) *model.PublicationEvent
}

type CPHMessageCreator struct {
}

func NewDefaultCPHMessageCreator() *CPHMessageCreator {
	return &CPHMessageCreator{}
}

func (cmc *CPHMessageCreator) ToPublicationEventMessage(coreAttributes *model.UppCoreContent, payload interface{}) (*producer.Message, error) {
	publicationEvent := cmc.ToPublicationEvent(coreAttributes, payload)

	jsonPublicationEvent, err := json.Marshal(publicationEvent)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-Request-Id":      coreAttributes.PublishReference,
		"Message-Timestamp": time.Now().Format(model.UPPDateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      "cms-content-published",
		"Content-Type":      "application/json",
		"Origin-System-Id":  model.MethodeSystemID,
	}

	return &producer.Message{Headers: headers, Body: string(jsonPublicationEvent)}, nil
}

func (cmc *CPHMessageCreator) ToPublicationEvent(coreAttributes *model.UppCoreContent, payload interface{}) *model.PublicationEvent {
	if coreAttributes.IsMarkedDeleted {
		return &model.PublicationEvent{
			ContentURI:   coreAttributes.ContentURI + coreAttributes.UUID,
			LastModified: coreAttributes.LastModified,
		}
	}
	return &model.PublicationEvent{
		ContentURI:   coreAttributes.ContentURI + coreAttributes.UUID,
		Payload:      payload,
		LastModified: coreAttributes.LastModified,
	}
}
