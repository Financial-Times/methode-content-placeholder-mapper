package handler

import (
	"testing"
	"github.com/stretchr/testify/mock"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"strings"
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
			UUID             : "512c1f3d-e48c-4618-863c-94bc9d913b9b",
			PublishReference : "tid_test123",
			LastModified     : "2017-05-15T15:54:32.166Z",
			ContentURI       : "",
			IsMarkedDeleted  : false,
		},
		&model.UppCoreContent{
			UUID             : "43dc1ff3-6d6c-41f3-9196-56dcaa554905",
			PublishReference : "tid_test123",
			LastModified     : "2017-05-15T15:54:32.166Z",
			ContentURI       : "",
			IsMarkedDeleted  : false,
		},
	}

	nativeMapper := new(mockNativeMapper)
	var nilErr *utility.MappingError
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nilErr)

	mockedAggregateCPHMapper := new(mockCPHAggregateMapper)
	mockedAggregateCPHMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), "tid_test123", "2017-05-15T15:54:32.166Z").Return(uppContents, nilErr)

	mockedMessageCreator := new(mockMessageCreator)
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "512c1f3d-e48c-4618-863c-94bc9d913b9b" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
				Body: "{\"uuid\":\"512c1f3d-e48c-4618-863c-94bc9d913b9b}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
				Headers: map[string]string {
					"X-Request-Id": "tid_test123",
				},
		}, nilErr)
	mockedMessageCreator.On("ToPublicationEventMessage", mock.MatchedBy(func(c *model.UppCoreContent) bool { return c.UUID == "43dc1ff3-6d6c-41f3-9196-56dcaa554905" }), mock.MatchedBy(func(p interface{}) bool { return true })).
		Return(&producer.Message{
		Body: "{\"uuid\":\"43dc1ff3-6d6c-41f3-9196-56dcaa554905}\",\"lastModifiedDate\":\"2017-05-15T15:54:32.166Z\"}",
		Headers: map[string]string {
			"X-Request-Id": "tid_test123",
		},
	}, nilErr)

	mockedProducer := new(mockProducer)
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

type mockCPHAggregateMapper struct {
	mock.Mock
}

func (m *mockCPHAggregateMapper) MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, tid, lmd string) ([]model.UppContent, *utility.MappingError) {
	args := m.Called(mpc, tid, lmd)
	return args.Get(0).([]model.UppContent), args.Get(1).(*utility.MappingError)
}

type mockProducer struct {
	mock.Mock
}

func (p *mockProducer) SendMessage(key string, msg producer.Message) error {
	args := p.Called(key, msg)
	return args.Error(0)
}

func (p *mockProducer) ConnectivityCheck() (string, error) {
	args := p.Called()
	return args.String(0), args.Error(1)
}

type mockMessageCreator struct {
	mock.Mock
}

func (m *mockMessageCreator) ToPublicationEventMessage(coreAttributes *model.UppCoreContent, payload interface{}) (*producer.Message, *utility.MappingError) {
	args := m.Called(coreAttributes, payload)
	return args.Get(0).(*producer.Message), args.Get(1).(*utility.MappingError)
}

func (m *mockMessageCreator) ToPublicationEvent(coreAttributes *model.UppCoreContent, payload interface{}) *model.PublicationEvent {
	args := m.Called(coreAttributes, payload)
	return args.Get(0).(*model.PublicationEvent)
}

type mockNativeMapper struct {
	mock.Mock
}

func (m *mockNativeMapper) Map(messageBody []byte) (*model.MethodeContentPlaceholder, *utility.MappingError) {
	args := m.Called(messageBody)
	return args.Get(0).(*model.MethodeContentPlaceholder), args.Get(1).(*utility.MappingError)
}
