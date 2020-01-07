package handler

import (
	"strings"
	"testing"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	consumer "github.com/Financial-Times/message-queue-gonsumer"
	"github.com/Financial-TimesFinancial-Times/methode-content-placeholder-mapper/v2/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

const methodeSystemOrigin = "http://cmdb.ft.com/systems/methode-web-pub"

func TestOnMessage_Ok(t *testing.T) {
	sourceMsg := consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      "tid_test123",
			"Origin-System-Id":  methodeSystemOrigin,
			"Message-Timestamp": "2017-05-15T15:54:32.166Z",
		},
		Body: "",
	}
	uppContents := []model.UppContent{
		&model.UppCoreContent{
			UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
		&model.UppCoreContent{
			UUID:             "43dc1ff3-6d6c-41f3-9196-56dcaa554905",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
	}

	nativeMapper := new(model.MockNativeMapper)
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nil)

	mockedAggregateCPHMapper := new(model.MockCPHAggregateMapper)
	mockedAggregateCPHMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), "tid_test123", "2017-05-15T15:54:32.166Z").Return(uppContents, nil)

	mockedMessageCreator := new(model.MockMessageCreator)
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "512c1f3d-e48c-4618-863c-94bc9d913b9b" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
			Body: "{\"uuid\":\"512c1f3d-e48c-4618-863c-94bc9d913b9b}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
			Headers: map[string]string{
				"X-Request-Id": "tid_test123",
			},
		}, nil)
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "43dc1ff3-6d6c-41f3-9196-56dcaa554905" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
			Body: "{\"uuid\":\"43dc1ff3-6d6c-41f3-9196-56dcaa554905}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
			Headers: map[string]string{
				"X-Request-Id": "tid_test123",
			},
		}, nil)

	mockedProducer := new(model.MockProducer)
	mockedProducer.On("SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true })).Return(nil)

	q := NewCPHMessageHandler(nil, mockedProducer, mockedAggregateCPHMapper, nativeMapper, mockedMessageCreator)
	q.HandleMessage(sourceMsg)

	mockedProducer.AssertCalled(t, "SendMessage", "",
		mock.MatchedBy(func(msg producer.Message) bool {
			return strings.Contains(msg.Body, "512c1f3d-e48c-4618-863c-94bc9d913b9b") && strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z")
		}))
	mockedProducer.AssertCalled(t, "SendMessage", "",
		mock.MatchedBy(func(msg producer.Message) bool {
			return strings.Contains(msg.Body, "43dc1ff3-6d6c-41f3-9196-56dcaa554905") && strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z")
		}))

	mockedProducer.AssertNumberOfCalls(t, "SendMessage", 2)
}

func TestOnMessageNativeMapError_MessagesNotSent(t *testing.T) {
	sourceMsg := consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      "tid_test123",
			"Origin-System-Id":  methodeSystemOrigin,
			"Message-Timestamp": "2017-05-15T15:54:32.166Z",
		},
		Body: "",
	}

	nativeMapper := new(model.MockNativeMapper)
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, errors.New("Some native mapping error"))

	mockedAggregateCPHMapper := new(model.MockCPHAggregateMapper)
	mockedMessageCreator := new(model.MockMessageCreator)
	mockedProducer := new(model.MockProducer)

	q := NewCPHMessageHandler(nil, mockedProducer, mockedAggregateCPHMapper, nativeMapper, mockedMessageCreator)
	q.HandleMessage(sourceMsg)

	mockedProducer.AssertNotCalled(t, "SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true }))
	mockedProducer.AssertNumberOfCalls(t, "SendMessage", 0)
}

func TestOnMessageCPHMapError_MessagesNotSent(t *testing.T) {
	sourceMsg := consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      "tid_test123",
			"Origin-System-Id":  methodeSystemOrigin,
			"Message-Timestamp": "2017-05-15T15:54:32.166Z",
		},
		Body: "",
	}
	uppContents := []model.UppContent{
		&model.UppCoreContent{
			UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
		&model.UppCoreContent{
			UUID:             "43dc1ff3-6d6c-41f3-9196-56dcaa554905",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
	}

	nativeMapper := new(model.MockNativeMapper)
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nil)

	mockedAggregateCPHMapper := new(model.MockCPHAggregateMapper)
	mockedAggregateCPHMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), "tid_test123", "2017-05-15T15:54:32.166Z").Return(uppContents, errors.New("Some cph mapping error"))

	mockedMessageCreator := new(model.MockMessageCreator)
	mockedProducer := new(model.MockProducer)

	q := NewCPHMessageHandler(nil, mockedProducer, mockedAggregateCPHMapper, nativeMapper, mockedMessageCreator)
	q.HandleMessage(sourceMsg)

	mockedProducer.AssertNotCalled(t, "SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true }))
	mockedProducer.AssertNumberOfCalls(t, "SendMessage", 0)
}

func TestOnMessagePublicationEventError_MessagesNotSent(t *testing.T) {
	sourceMsg := consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      "tid_test123",
			"Origin-System-Id":  methodeSystemOrigin,
			"Message-Timestamp": "2017-05-15T15:54:32.166Z",
		},
		Body: "",
	}
	uppContents := []model.UppContent{
		&model.UppCoreContent{
			UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
		&model.UppCoreContent{
			UUID:             "43dc1ff3-6d6c-41f3-9196-56dcaa554905",
			PublishReference: "tid_test123",
			LastModified:     "2017-05-15T15:54:32.166Z",
			ContentURI:       "",
			IsMarkedDeleted:  false,
		},
	}

	nativeMapper := new(model.MockNativeMapper)
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nil)

	mockedAggregateCPHMapper := new(model.MockCPHAggregateMapper)
	mockedAggregateCPHMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), "tid_test123", "2017-05-15T15:54:32.166Z").Return(uppContents, nil)

	mockedMessageCreator := new(model.MockMessageCreator)
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "512c1f3d-e48c-4618-863c-94bc9d913b9b" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
			Body: "{\"uuid\":\"512c1f3d-e48c-4618-863c-94bc9d913b9b}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
			Headers: map[string]string{
				"X-Request-Id": "tid_test123",
			},
		}, errors.New("Error creating publication event messages."))
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "43dc1ff3-6d6c-41f3-9196-56dcaa554905" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
			Body: "{\"uuid\":\"43dc1ff3-6d6c-41f3-9196-56dcaa554905}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
			Headers: map[string]string{
				"X-Request-Id": "tid_test123",
			},
		}, errors.New("Error creating publication event messages."))

	mockedProducer := new(model.MockProducer)
	mockedProducer.On("SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true })).Return(nil)

	q := NewCPHMessageHandler(nil, mockedProducer, mockedAggregateCPHMapper, nativeMapper, mockedMessageCreator)
	q.HandleMessage(sourceMsg)

	mockedProducer.AssertNotCalled(t, "SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true }))
	mockedProducer.AssertNumberOfCalls(t, "SendMessage", 0)
}
