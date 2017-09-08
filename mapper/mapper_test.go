package mapper

import (
	"errors"
	"io/ioutil"
	"regexp"
	"testing"
	"time"

	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

const expectedTransactionID = "tid_bh7VTFj9Il"

var uuidRegexp = regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")

func TestCorrectMappingToUpdateEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderUpdateMsg()
	expectedCPHPubEventMsg := buildIgPlaceholderPubEvent()
	expectedCCPubEventMsg := buildIgComplementaryContentPubEvent()

	mapper := NewDefaultMapper()

	actualCPHPubEventMsg, _, actualCCPubEventMsg, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "It should not return error in mapping placeholder")

	verifyMappingIsCorrect(t, actualCPHPubEventMsg, &expectedCPHPubEventMsg)
	verifyMappingIsCorrect(t, actualCCPubEventMsg, &expectedCCPubEventMsg)
}

func verifyMappingIsCorrect(t *testing.T, actualPubEventMsg *producer.Message, expectedPubEventMsg *producer.Message) {
	assert.Equal(t, expectedTransactionID, actualPubEventMsg.Headers["X-Request-Id"], "The Transaction ID should be consistent")
	assert.Equal(t, "cms-content-published", actualPubEventMsg.Headers["Message-Type"], "The Message type should be cms-content-published")
	assert.Equal(t, "application/json", actualPubEventMsg.Headers["Content-Type"], "The Content type should be application/json")
	assert.Regexp(t, uuidRegexp, actualPubEventMsg.Headers["Message-Id"], "The Message ID should be a valid UUID")
	_, parseErr := time.Parse(model.UPPDateFormat, actualPubEventMsg.Headers["Message-Timestamp"])
	assert.Nil(t, parseErr, "The message timestamp should have a consistent format")

	expectedBodyAsMap := jsonStringToMap(expectedPubEventMsg.Body, t)
	actualBodyAsMap := jsonStringToMap(actualPubEventMsg.Body, t)
	assert.Equal(t, expectedBodyAsMap, actualBodyAsMap, "The placeholder should be mapped properly")
}

func buildIgMethodePlaceholderUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_update.json")
}

func buildIgPlaceholderPubEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_placeholder_pub_event.json")
}

func buildIgComplementaryContentPubEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_complementarycontent_pub_event.json")
}

func TestCorrectMappingToDeleteEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderDeleteMsg()
	expectedCPHPubEventMsg := buildIgPlaceholderDeleteEvent()
	expectedCCPubEventMsg := buildIgComplementaryContentDeleteEvent()

	mapper := NewDefaultMapper()

	actualCPHPubEventMsg, _, actualCCPubEventMsg, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "It should not return error in mapping placeholder")

	verifyMappingIsCorrect(t, actualCPHPubEventMsg, &expectedCPHPubEventMsg)
	verifyMappingIsCorrect(t, actualCCPubEventMsg, &expectedCCPubEventMsg)
}

func buildIgMethodePlaceholderDeleteMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_delete.json")
}

func buildIgPlaceholderDeleteEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_placeholder_delete_event.json")
}

func buildIgComplementaryContentDeleteEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_complementarycontent_delete_event.json")
}

func TestCorrectMappingToUpdateEventWithHeadlineOnly(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderOnlyHeadlineUpdateMsg()
	expectedCPHPubEventMsg := buildIgPlaceholderOnlyHeadlinePubEvent()

	mapper := NewDefaultMapper()

	actualCPHPubEventMsg, _, _, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "It should not return error in mapping placeholder")

	verifyMappingIsCorrect(t, actualCPHPubEventMsg, &expectedCPHPubEventMsg)
}

func buildIgMethodePlaceholderOnlyHeadlineUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_headline_only.json")
}

func buildIgPlaceholderOnlyHeadlinePubEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_placeholder_headline_only_pub_event.json")
}

func TestHandleMethodePlaceholderEvent(t *testing.T) {
	producerMock := new(QueueProducerMock)
	producerMock.On("SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message")).Return(nil)

	mapper := NewDefaultMapper()
	mapper.messageProducer = producerMock

	methodeMsg := buildIgMethodePlaceholderUpdateMsg()
	mapper.HandlePlaceholderMessages(methodeMsg)

	producerMock.AssertCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func TestDoNotMapMethodeArticleDeleteEvent(t *testing.T) {
	methodeArticleMsg := buildMethodeArticleDeleteMsg()
	mapper := NewDefaultMapper()

	_, _, _, _, err := mapper.mapMessage(methodeArticleMsg)
	assert.EqualError(t, err, "Methode content is not a content placeholder", "The mapping of the article should be unsuccessful")
}

func TestDoNotHandleMethodeArticleDeleteEvent(t *testing.T) {
	producerMock := new(QueueProducerMock)

	mapper := NewDefaultMapper()
	mapper.messageProducer = producerMock

	methodeArticleMsg := buildMethodeArticleDeleteMsg()
	mapper.HandlePlaceholderMessages(methodeArticleMsg)

	producerMock.AssertNotCalled(t, "SendMessage")
}

func TestNotHandleMethodePlaceholderEventWhenProducerReturnsError(t *testing.T) {
	producerMock := new(QueueProducerMock)
	producerMock.On("SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message")).Return(errors.New("I do not want to send the message! I'm on strike!"))

	mapper := NewDefaultMapper()
	mapper.messageProducer = producerMock

	methodeMsg := buildIgMethodePlaceholderUpdateMsg()
	mapper.HandlePlaceholderMessages(methodeMsg)

	producerMock.AssertCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func buildMethodeArticleDeleteMsg() consumer.Message {
	return buildMethodeMsg("test_resources/methode_article_delete.json")
}

func TestDoNotMapPlaceholderWithNoURLInHeadline(t *testing.T) {
	placeholderMsg := buildIgMethodePlaceholderNoURLUpdateMsg()
	mapper := NewDefaultMapper()

	_, _, _, _, err := mapper.mapMessage(placeholderMsg)
	assert.EqualError(t, err, "Methode Content headline does not contain a link", "The mapping of the placeholder should be unsuccessful")
}

func buildIgMethodePlaceholderNoURLUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_no_url.json")
}

func TestDoNotMapPlaceholderWithWrongURLInHeadline(t *testing.T) {
	placeholderMsg := buildIgMethodePlaceholderWithWrongURLUpdateMsg()
	mapper := NewDefaultMapper()

	_, _, _, _, err := mapper.mapMessage(placeholderMsg)
	assert.EqualError(t, err, "Methode Content headline does not contain a valid URL - parse %gh&%ij: invalid URL escape \"%gh\"", "The mapping of the placeholder should be unsuccessful")
}

func buildIgMethodePlaceholderWithWrongURLUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_wrong_url.json")
}

func TestDoNotMapPlaceholderWithRelativeURLInHeadline(t *testing.T) {
	placeholderMsg := buildIgMethodePlaceholderWithRelativeURLUpdateMsg()
	mapper := NewDefaultMapper()

	_, _, _, _, err := mapper.mapMessage(placeholderMsg)
	assert.EqualError(t, err, "Methode Content headline does not contain an absolute URL", "The mapping of the placeholder should be unsuccessful")
}

func buildIgMethodePlaceholderWithRelativeURLUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_relative_url.json")
}

func TestDoNotHandleBrightcoveVideoEvent(t *testing.T) {
	producerMock := new(QueueProducerMock)

	mapper := NewDefaultMapper()
	mapper.messageProducer = producerMock

	videoMsg := buildBrightcoveVideoMsg()
	mapper.HandlePlaceholderMessages(videoMsg)

	producerMock.AssertNotCalled(t, "SendMessage")
}

func TestDummyCpHeadlineIsIgnored(t *testing.T) {
	methodeMsg := buildMethodeMsg("test_resources/ig_methode_placeholder_dummy_cp_title.json")
	m := NewDefaultMapper()

	cphMessage, _, ccMessage, _, err := m.mapMessage(methodeMsg)
	assert.Nil(t, err)

	cphBodyMap := jsonStringToMap(cphMessage.Body, t)
	ccBodyMap := jsonStringToMap(ccMessage.Body, t)

	cphPayload, ok := cphBodyMap["payload"]
	assert.True(t, ok)
	ccPayload, ok := ccBodyMap["payload"]
	assert.True(t, ok)

	cphPayloadMap, ok := cphPayload.(map[string]interface{})
	assert.True(t, ok)
	ccPayloadMap, ok := ccPayload.(map[string]interface{})
	assert.True(t, ok)

	alternativeTitles, ok := cphPayloadMap["alternativeTitles"]
	assert.True(t, ok)
	assert.Nil(t, alternativeTitles)

	promotionalTitle, ok := ccPayloadMap["promotionalTitle"]
	assert.True(t, ok)
	assert.NotNil(t, promotionalTitle)
}

func buildBrightcoveVideoMsg() consumer.Message {
	return consumer.Message{
		Body: "",
		Headers: map[string]string{
			"Origin-System-Id":  "http://cmdb.ft.com/systems/brightcove",
			"X-Request-Id":      expectedTransactionID,
			"Message-Timestamp": "2016-12-16T13:13:51.154Z",
		},
	}
}

func TestNotMappingWithWrongHeadline(t *testing.T) {
	badPlaceholderMsg := buildMethodePlaceholderBadHeadlineMsg()
	mapper := NewDefaultMapper()

	_, _, _, _, err := mapper.mapMessage(badPlaceholderMsg)

	assert.EqualError(t, err, "Methode Content headline does not contain text", "The mapping of the placeholder should be unsuccessful")
}

func buildMethodePlaceholderBadHeadlineMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_bad_headline.json")
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

func buildProducerMessage(filePath string) producer.Message {
	pubEventBody, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return producer.Message{
		Body: string(pubEventBody),
		Headers: map[string]string{
			"X-Request-Id":      expectedTransactionID,
			"Message-Timestamp": "2016-12-16T13:13:51.154Z",
		},
	}
}

type QueueProducerMock struct {
	mock.Mock
}

func (p *QueueProducerMock) SendMessage(s string, msg producer.Message) error {
	args := p.Called(s, msg)
	return args.Error(0)
}

func (*QueueProducerMock) ConnectivityCheck() (string, error) {
	return "OK", nil
}

func jsonStringToMap(marshalled string, t *testing.T) map[string]interface{} {
	var unmarshalled map[string]interface{}
	err := json.Unmarshal([]byte(marshalled), &unmarshalled)
	assert.NoError(t, err, "Unmashalling the json content has encountered and error")
	return unmarshalled
}
