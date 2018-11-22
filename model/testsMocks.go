package model

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/stretchr/testify/mock"
)

type MockNativeMapper struct {
	mock.Mock
}

func (m *MockNativeMapper) Map(messageBody []byte) (*MethodeContentPlaceholder, error) {
	args := m.Called(messageBody)
	return args.Get(0).(*MethodeContentPlaceholder), args.Error(1)
}

type MockCPHAggregateMapper struct {
	mock.Mock
}

func (m *MockCPHAggregateMapper) MapContentPlaceholder(mpc *MethodeContentPlaceholder, tid, lmd string) ([]UppContent, error) {
	args := m.Called(mpc, tid, lmd)
	return args.Get(0).([]UppContent), args.Error(1)
}

type MockProducer struct {
	mock.Mock
}

func (p *MockProducer) SendMessage(key string, msg producer.Message) error {
	args := p.Called(key, msg)
	return args.Error(0)
}

func (p *MockProducer) ConnectivityCheck() (string, error) {
	args := p.Called()
	return args.String(0), args.Error(1)
}

type MockMessageCreator struct {
	mock.Mock
}

func (m *MockMessageCreator) ToPublicationEventMessage(coreAttributes *UppCoreContent, payload interface{}) (*producer.Message, error) {
	args := m.Called(coreAttributes, payload)
	return args.Get(0).(*producer.Message), args.Error(1)
}

func (m *MockMessageCreator) ToPublicationEvent(coreAttributes *UppCoreContent, payload interface{}) *PublicationEvent {
	args := m.Called(coreAttributes, payload)
	return args.Get(0).(*PublicationEvent)
}

type MockDocStoreClient struct {
	mock.Mock
}

func (m *MockDocStoreClient) ContentQuery(authority string, identifier string, tid string) (status int, location string, err error) {
	args := m.Called(authority, identifier, tid)
	return args.Int(0), args.String(1), args.Error(2)
}

func (m *MockDocStoreClient) GetContent(uuid, tid string) (*DocStoreUppContent, error) {
	args := m.Called(uuid, tid)
	return args.Get(0).(*DocStoreUppContent), args.Error(1)
}

func (m *MockDocStoreClient) ConnectivityCheck() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDocStoreClient) CheckContentExists(uuid, tid string) error {
	args := m.Called(uuid, tid)
	return args.Error(0)
}

type MockIResolver struct {
	mock.Mock
}

func (m *MockIResolver) CheckContentExists(uuid, tid string) error {
	args := m.Called(uuid, tid)
	return args.Error(0)
}

func (m *MockIResolver) ResolveIdentifier(serviceId, refField, tid string) (string, error) {
	args := m.Called(serviceId, refField, tid)
	return args.String(0), args.Error(1)
}

type MockCPHMapper struct {
	mock.Mock
}

func (m *MockCPHMapper) MapContentPlaceholder(mpc *MethodeContentPlaceholder, uuid, tid, lmd string) ([]UppContent, error) {
	args := m.Called(mpc, uuid, tid, lmd)
	return args.Get(0).([]UppContent), args.Error(1)
}

type MockCPHValidator struct {
	mock.Mock
}

func (m *MockCPHValidator) Validate(mcp *MethodeContentPlaceholder) error {
	args := m.Called(mcp)
	return args.Error(0)
}
