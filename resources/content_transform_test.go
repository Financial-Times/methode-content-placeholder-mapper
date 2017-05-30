package resources

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
)

const placeholderMsg = `{"foo":"bar"}`
const mapperURL = "http://methode-content-placeholder-mapper/map"

func TestSuccessfulMapEndpoint(t *testing.T) {
	placeholder := mapper.UpContentPlaceholder{WebURL: "http://www.ft.com/ig/sites/2014/virgingroup-timeline/"}
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, (*mapper.MappingError)(nil))
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(placeholder, (*mapper.MappingError)(nil))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "It should return status 200")
}

func TestDeletedContentPlaceholderMapEndpoint(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, (*mapper.MappingError)(nil))
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(mapper.UpContentPlaceholder{}, (*mapper.MappingError)(nil))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "It should return status 404")
	assert.Equal(t, "text/plain" , w.Header().Get("Content-Type"), "The Content-Type header should be text/plain")
	assert.NotEmpty(t, w.Body.Bytes(), "The response body should not be empty")
}

func TestUnsuccessfulMethodePlaceholderBuild(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, mapper.NewMappingError().WithMessage("What is it?"))
	h := NewMapEndpointHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
	assert.Equal(t, "What is it?\n", w.Body.String())
}

func TestUnsuccessfulPlaceholderMapping(t *testing.T) {
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

func (m *MapperMock) StartMappingMessages(c consumer.MessageConsumer, p producer.MessageProducer) {
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
