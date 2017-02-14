package resources

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const placeholderMsg = `{"foo":"bar"}`
const mapperURL = "http://example.com/content-transform/2eb712b6-70bf-4f18-a958-cd99bcc20ad2"

func TestSucessfulMapEndpoint(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, (*mapper.MappingError)(nil))
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(mapper.UpContentPlaceholder{}, (*mapper.MappingError)(nil))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, w.Code, http.StatusOK, "It should return status 200")
}

func TestUnsucessfulMethodePlaceholderBuild(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, mapper.NewMappingError().WithMessage("What is it?"))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
	assert.Equal(t, "What is it?\n", w.Body.String())
}

func TestUnsucessfulPlaceholderMapping(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, (*mapper.MappingError)(nil))
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(mapper.UpContentPlaceholder{}, mapper.NewMappingError().WithMessage("All map and no play makes MCPM a dull boy"))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
	assert.Equal(t, "All map and no play makes MCPM a dull boy\n", w.Body.String())
}

type MapperMock struct {
	mock.Mock
}

func (m *MapperMock) HandlePlaceholderMessages(msg consumer.Message) {
	m.Called(msg)
}

func (m *MapperMock) StartMappingMessages(c consumer.Consumer, p producer.MessageProducer) {
	m.Called(c, p)
}

func (m *MapperMock) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (mapper.MethodeContentPlaceholder, *mapper.MappingError) {
	args := m.Called(r)
	return args.Get(0).(mapper.MethodeContentPlaceholder), args.Get(1).(*mapper.MappingError)
}

func (m *MapperMock) MapContentPlaceholder(mpc mapper.MethodeContentPlaceholder) (mapper.UpContentPlaceholder, *mapper.MappingError) {
	args := m.Called(mpc)
	return args.Get(0).(mapper.UpContentPlaceholder), args.Get(1).(*mapper.MappingError)
}
