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

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"io/ioutil"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
)

const placeholderMsg = `{"uuid":"f9845f8a-c210-11e6-91a7-e73ace06f770", "type": "EOM::CompoundStory"}`
const mapperURL = "http://methode-content-placeholder-mapper/map"
const expectedTransactionID = "tid_bh7VTFj9Il"

func TestSuccessfulMapEndpoint(t *testing.T) {
	methodeContentMsg := buildIgMethodePlaceholderUpdateMsg()
	mockedResolver := new(mockResolver)
	cphValidator := mapper.NewDefaultCPHValidator()

	aggregateMapper := mapper.NewAggregateCPHMapper(mockedResolver, cphValidator, )
	messageCreator := message.NewDefaultCPHMessageCreator()
	nativeMapper := mapper.DefaultMessageMapper{}

	h := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(methodeContentMsg.Body)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "It should return status 200")
}

func TestDeletedContentPlaceholderMapEndpoint(t *testing.T) {
	methodeContentDeleteMsg := buildIgMethodePlaceholderDeleteMsg()

	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := message.NewDefaultCPHMessageCreator()
	nativeMapper := mapper.DefaultMessageMapper{}

	h := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(methodeContentDeleteMsg.Body)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "It should return status 404")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "The Content-Type header should be application/json")
	assert.NotEmpty(t, w.Body.Bytes(), "The response body should not be empty")
}

func TestUnsuccessfulMethodePlaceholderBuild(t *testing.T) {
	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := message.NewDefaultCPHMessageCreator()
	nativeMapper := mapper.DefaultMessageMapper{}

	h := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(nil)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
}

func TestUnsuccessfulPlaceholderMapping(t *testing.T) {
	m := new(MapperMock)
	m.On("MapContentPlaceholder", mock.Anything).Return(model.UppContentPlaceholder{}, model.UppComplementaryContent{}, utility.NewMappingError().WithMessage("All map and no play makes MCPM a dull boy"))

	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := message.NewDefaultCPHMessageCreator()
	nativeMapper := mapper.DefaultMessageMapper{}

	h := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(placeholderMsg)))
	w := httptest.NewRecorder()
	h.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
}

func TestSuccesfulBuildOfPlaceholderFromHTTPRequest(t *testing.T) {
	placeholderBody, err := ioutil.ReadFile("../mapper/test_resources/ig_methode_placeholder_update.json")
	if err != nil {
		panic(err)
	}

	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := message.NewDefaultCPHMessageCreator()
	nativeMapper := mapper.DefaultMessageMapper{}

	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(placeholderBody))
	mapHandler := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	methodePlacheholder, err := mapHandler.NewMethodeContentPlaceholderFromHTTPRequest(req)
	assert.Nil(t, err, "It should not return an error")
	assert.NotZero(t, methodePlacheholder, "pippo")
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

func (m *MapperMock) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder) (*model.UppContentPlaceholder, *model.UppComplementaryContent, *utility.MappingError) {
	args := m.Called(mcp)
	return args.Get(0).(*model.UppContentPlaceholder), args.Get(1).(*model.UppComplementaryContent), args.Get(2).(*utility.MappingError)
}

func buildIgMethodePlaceholderUpdateMsg() consumer.Message {
	return buildMethodeMsg("../mapper/test_resources/ig_methode_placeholder_update.json")
}

func buildIgMethodePlaceholderDeleteMsg() consumer.Message {
	return buildMethodeMsg("../mapper/test_resources/ig_methode_placeholder_delete.json")
}

func buildMethodeMsg(examplePath string) consumer.Message {
	placeholderBody, err := ioutil.ReadFile(examplePath)
	if err != nil {
		panic(err)
	}
	return consumer.Message{
		Body: string(placeholderBody),
		Headers: map[string]string{
			"Origin-System-Id":  model.MethodeSystemID,
			"X-Request-Id":      expectedTransactionID,
			"Message-Timestamp": "2016-12-16T13:13:51.154Z",
		},
	}
}
