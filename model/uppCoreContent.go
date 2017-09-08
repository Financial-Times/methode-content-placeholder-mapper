package model

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"time"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

type UppContent interface {
	ToPublicationEventMessage() (*producer.Message, *utility.MappingError)
}

type UppCoreContent struct {
	UUID             string             `json:"uuid"`
	PublishReference string             `json:"publishReference"`
	LastModified     string             `json:"lastModified"`
	ContentURI       string             `json:"-"`
	IsMarkedDeleted  bool               `json:"-"`
}

type PublicationEvent struct {
	ContentURI   string                `json:"contentUri"`
	Payload      interface{}           `json:"payload,omitempty"`
	LastModified string                `json:"lastModified"`
}

func (buc *UppCoreContent) ToPublicationEventMessage(payload interface{}) (*producer.Message, *utility.MappingError) {
	publicationEvent := buc.toPublicationEvent(payload)

	jsonPublicationEvent, err := json.Marshal(publicationEvent)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(buc.UUID)
	}

	headers := map[string]string{
		"X-Request-Id":      buc.PublishReference,
		"Message-Timestamp": time.Now().Format(UPPDateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      "cms-content-published",
		"Content-Type":      "application/json",
		"Origin-System-Id":  MethodeSystemID,
	}

	return &producer.Message{Headers: headers, Body: string(jsonPublicationEvent)}, nil
}

func (buc *UppCoreContent) toPublicationEvent(payload interface{}) *PublicationEvent {
	if buc.IsMarkedDeleted {
		return &PublicationEvent{
			ContentURI:   buc.ContentURI + buc.UUID,
			LastModified: buc.LastModified,
		}
	}
	return &PublicationEvent{
		ContentURI:   buc.ContentURI + buc.UUID,
		Payload:      payload,
		LastModified: buc.LastModified,
	}
}
