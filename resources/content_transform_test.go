package resources

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"io/ioutil"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
	"github.com/stretchr/testify/mock"
	"fmt"
	"encoding/json"
)

const mapperURL = "http://methode-content-placeholder-mapper/map"
const expectedTransactionID = "tid_bh7VTFj9Il"

func TestMapEndpoint_Ok(t *testing.T) {
	methodeContentMsg := buildIgMethodePlaceholderUpdateMsg()

	aggregateMapper := new(model.MockCPHAggregateMapper)
	nativeMapper := new(model.MockNativeMapper)
	messageCreator := message.NewDefaultCPHMessageCreator()

	uppContent := []model.UppContent{
		&model.UppContentPlaceholder{},
		&model.UppComplementaryContent{},
	}

	nativeMapper.On("Map", mock.MatchedBy(func([]byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nil)
	aggregateMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), mock.MatchedBy(func(string) bool { return true }), mock.MatchedBy(func(string) bool { return true })).Return(uppContent, nil)

	mapHandler := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(methodeContentMsg.Body)))
	w := httptest.NewRecorder()
	mapHandler.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "It should return status 200")
}

func TestMapEndpointFailedNativeTransformation_Returns422(t *testing.T) {
	aggregateMapper := new(model.MockCPHAggregateMapper)
	nativeMapper := new(model.MockNativeMapper)
	messageCreator := message.NewDefaultCPHMessageCreator()

	uppContent := []model.UppContent{
		&model.UppContentPlaceholder{},
		&model.UppComplementaryContent{},
	}

	nativeMapper.On("Map", mock.MatchedBy(func([]byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, fmt.Errorf("Error decoding or unmarshalling methode body."))
	aggregateMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), mock.MatchedBy(func(string) bool { return true }), mock.MatchedBy(func(string) bool { return true })).Return(uppContent, nil)

	mapHandler := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(nil)))
	w := httptest.NewRecorder()
	mapHandler.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
}

func TestMapEndpointDeleteMessage_Returns404(t *testing.T) {
	aggregateMapper := new(model.MockCPHAggregateMapper)
	nativeMapper := new(model.MockNativeMapper)
	messageCreator := message.NewDefaultCPHMessageCreator()

	uppContent := []model.UppContent{
		&model.UppContentPlaceholder{},
		&model.UppComplementaryContent{},
	}

	methodeContent := &model.MethodeContentPlaceholder{
		Attributes: model.Attributes{
			IsDeleted: true,
		},
	}

	expectedMessage, _ := json.Marshal(&msg{Message: "Delete event"})

	nativeMapper.On("Map", mock.MatchedBy(func([]byte) bool { return true })).Return(methodeContent, nil)
	aggregateMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }), mock.MatchedBy(func(string) bool { return true }), mock.MatchedBy(func(string) bool { return true })).Return(uppContent, nil)

	mapHandler := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(nil)))
	w := httptest.NewRecorder()
	mapHandler.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "It should return status 404")
	assert.Equal(t, expectedMessage, w.Body.Bytes(), "It should send delete message")
}

func TestMapEndpointFailedTransformation_Returns422(t *testing.T) {
	aggregateMapper := new(model.MockCPHAggregateMapper)
	nativeMapper := new(model.MockNativeMapper)
	messageCreator := message.NewDefaultCPHMessageCreator()

	uppContent := []model.UppContent{
		&model.UppContentPlaceholder{},
		&model.UppComplementaryContent{},
	}

	nativeMapper.On("Map", mock.MatchedBy(func([]byte) bool { return true })).Return(&model.MethodeContentPlaceholder{}, nil)
	aggregateMapper.On("MapContentPlaceholder", mock.MatchedBy(func(mpc *model.MethodeContentPlaceholder) bool { return true }),
		mock.MatchedBy(func(string) bool { return true }),
		mock.MatchedBy(func(string) bool { return true })).Return(uppContent, fmt.Errorf("Error transforming model."))

	mapHandler := NewMapEndpointHandler(aggregateMapper, messageCreator, nativeMapper)

	req := httptest.NewRequest("POST", mapperURL, bytes.NewReader([]byte(nil)))
	w := httptest.NewRecorder()
	mapHandler.ServeMapEndpoint(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "It should return status 422")
}

func buildIgMethodePlaceholderUpdateMsg() consumer.Message {
	return buildMethodeMsg("../mapper/test_resources/methode_cph_update.json")
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
