package resources

import (
	"bytes"
	"errors"
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

func TestSucessfulContentTransformation(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, nil)
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(mapper.UpContentPlaceholder{}, nil)
	h := NewContentTransformHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK, "It should return status 200")
}

func TestUnsucessfulMethodePlaceholderBuild(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, errors.New("What is it?"))
	h := NewContentTransformHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
	assert.Equal(t, "What is it?\n", w.Body.String())
}

func TestUnsucessfulPlaceholderMapping(t *testing.T) {
	m := new(MapperMock)
	m.On("NewMethodeContentPlaceholderFromHTTPRequest", mock.AnythingOfType("*http.Request")).Return(mapper.MethodeContentPlaceholder{}, nil)
	m.On("MapContentPlaceholder", mock.AnythingOfType("mapper.MethodeContentPlaceholder")).Return(mapper.UpContentPlaceholder{}, errors.New("All map and no play makes MCPM a dull boy"))
	h := NewContentTransformHandler(m)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

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

func (m *MapperMock) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (mapper.MethodeContentPlaceholder, error) {
	args := m.Called(r)
	return args.Get(0).(mapper.MethodeContentPlaceholder), args.Error(1)
}

func (m *MapperMock) MapContentPlaceholder(mpc mapper.MethodeContentPlaceholder) (mapper.UpContentPlaceholder, error) {
	args := m.Called(mpc)
	return args.Get(0).(mapper.UpContentPlaceholder), args.Error(1)
}
