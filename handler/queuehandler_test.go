package handler

import (
	"testing"
	"github.com/stretchr/testify/mock"
	"strings"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Financial-Times/message-queue-go-producer/producer"
)

const methodeSystemOrigin = "http://cmdb.ft.com/systems/methode-web-pub"

func TestOnMessage_Ok(t *testing.T) {
	sourceMsg := consumer.Message{
		Headers: map[string]string{
			"X-Request-Id":      "tid_test123",
			"Origin-System-Id":  methodeSystemOrigin,
			"Message-Timestamp": "2017-05-15T15:54:32.166Z",
		},
	}
	uppContents := []model.UppContent{
		model.UppCoreContent{
			UUID             : "512c1f3d-e48c-4618-863c-94bc9d913b9b",
			PublishReference : "",
			LastModified     : "2017-05-15T15:54:32.166Z",
			ContentURI       : "",
			IsMarkedDeleted  : false,
		},
		model.UppCoreContent{
			UUID             : "43dc1ff3-6d6c-41f3-9196-56dcaa554905",
			PublishReference : "",
			LastModified     : "2017-05-15T15:54:32.166Z",
			ContentURI       : "",
			IsMarkedDeleted  : false,
		},
	}
	mockedAggregateCPHMapper := new(mockAggregateCPHMapper)
	mockedAggregateCPHMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true })).Return(uppContents, nil)

	nativeMapper := new(mockNativeMapper)
	nativeMapper.On("Map", mock.MatchedBy(func(messageBody []byte, transactionID string, lastModified string) bool { return true })).Return(uppContents, nil)

	mockedProducer := new(mockProducer)
	mockedProducer.On("SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return true })).Return(nil)

	mockedMessageCreator := new(mockMessageCreator)

	q := NewCPHMessageHandler(nil, mockedProducer, mockedAggregateCPHMapper, nati, mockedMessageCreator)
	q.HandleMessage(sourceMsg)

	mockedProducer.AssertCalled(t, "SendMessage", "",
		mock.MatchedBy(func(msg producer.Message) bool {
			return strings.Contains(msg.Body, "512c1f3d-e48c-4618-863c-94bc9d913b9b") && strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z")
		}))
	mockedProducer.AssertCalled(t, "SendMessage", "",
		mock.MatchedBy(func(msg producer.Message) bool {
			return strings.Contains(msg.Body, "43dc1ff3-6d6c-41f3-9196-56dcaa554905") && strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z")
		}))
	mockedProducer.AssertCalled(t, "SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z") }))
	mockedProducer.AssertCalled(t, "SendMessage", "", mock.MatchedBy(func(msg producer.Message) bool { return strings.Contains(msg.Body, "2017-05-15T15:54:32.166Z") }))
	mockedProducer.AssertNumberOfCalls(t, "SendMessage", 2)
}

type mockAggregateCPHMapper struct {
	mock.Mock
}

func (m *mockAggregateCPHMapper) MapContentPlaceholder(mpc *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError) {
	args := m.Called(mpc)
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

func (m *mockNativeMapper) Map(messageBody []byte, transactionID string, lastModified string) (*model.MethodeContentPlaceholder, *utility.MappingError) {
	args := m.Called(messageBody, transactionID, lastModified)
	return args.Get(0).(*model.MethodeContentPlaceholder), args.Get(1).(*utility.MappingError)
}
