package mapper

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const expectedTransactionID = "tid_i1ktygkniy"

var uuidRegexp = regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")

func TestCorrectMappingToUpdateEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderUpdateMsg()
	expectedPubEventMsg := buildIgPlaceholderPubEvent()

	mapper := &mapper{}

	actualPubEventMsg, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "It should not return error in mapping placeholder")
	assert.Equal(t, expectedPubEventMsg.Body, actualPubEventMsg.Body, "The placeholder should be mapped properly")
	assert.Equal(t, expectedTransactionID, actualPubEventMsg.Headers["X-Request-Id"], "The Transaction ID should be consistent")
	assert.Equal(t, "cms-content-published", actualPubEventMsg.Headers["Message-Type"], "The Message type should be cms-content-published")
	assert.Equal(t, "application/json", actualPubEventMsg.Headers["Content-Type"], "The Content type should be application/json")
	assert.Regexp(t, uuidRegexp, actualPubEventMsg.Headers["Message-Id"], "The Message ID should be a valid UUID")
	_, parseErr := time.Parse(upDateFormat, actualPubEventMsg.Headers["Message-Timestamp"])
	assert.Nil(t, parseErr, "The message timestamp should have a consistent format")
}

func buildIgMethodePlaceholderUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_update.json")
}

func buildIgPlaceholderPubEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_placeholder_pub_event.json")
}

func TestCorrectMappingToDeleteEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderDeleteMsg()
	expectedPubEventMsg := buildIgPlaceholderDeleteEvent()

	mapper := &mapper{}

	actualPubEventMsg, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "It should not return error in mapping placeholder")
	assert.Equal(t, expectedPubEventMsg.Body, actualPubEventMsg.Body, "The placeholder should be mapped properly")
	assert.Equal(t, expectedTransactionID, actualPubEventMsg.Headers["X-Request-Id"], "The Transaction ID should be consistent")
	assert.Equal(t, "cms-content-published", actualPubEventMsg.Headers["Message-Type"], "The Message type should be cms-content-published")
	assert.Equal(t, "application/json", actualPubEventMsg.Headers["Content-Type"], "The Content type should be application/json")
	assert.Regexp(t, uuidRegexp, actualPubEventMsg.Headers["Message-Id"], "The Message ID should be a valid UUID")
	_, parseErr := time.Parse(upDateFormat, actualPubEventMsg.Headers["Message-Timestamp"])
	assert.Nil(t, parseErr, "The message timestamp should have a consistent format")
}

func buildIgMethodePlaceholderDeleteMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_delete.json")
}

func buildIgPlaceholderDeleteEvent() producer.Message {
	return buildProducerMessage("test_resources/ig_placeholder_delete_event.json")
}

func TestCorrectMappingToUpdateEventWithHeadlineOnly(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderOnlyHeadlineUpdateMsg()

	expectedPubEventMsg := buildIgPlaceholderOnlyHeadlinePubEvent()

	mapper := &mapper{}

	actualPubEventMsg, _, err := mapper.mapMessage(igMethodePlaceHolderMsg)

	assert.Nil(t, err, "It should not return error in mapping placeholder")
	assert.Equal(t, expectedPubEventMsg.Body, actualPubEventMsg.Body, "The placeholder should be mapped properly")
	assert.Equal(t, expectedTransactionID, actualPubEventMsg.Headers["X-Request-Id"], "The Transaction ID should be consistent")
	assert.Equal(t, "cms-content-published", actualPubEventMsg.Headers["Message-Type"], "The Message type should be cms-content-published")
	assert.Equal(t, "application/json", actualPubEventMsg.Headers["Content-Type"], "The Content type should be application/json")
	assert.Regexp(t, uuidRegexp, actualPubEventMsg.Headers["Message-Id"], "The Message ID should be a valid UUID")
	_, parseErr := time.Parse(upDateFormat, actualPubEventMsg.Headers["Message-Timestamp"])
	assert.Nil(t, parseErr, "The message timestamp should have a consistent format")
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

	mapper := &mapper{messageProducer: producerMock}

	methodeMsg := buildIgMethodePlaceholderUpdateMsg()
	mapper.HandlePlaceholderMessages(methodeMsg)

	producerMock.AssertCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func TestDoNotMapMethodeArticleDeleteEvent(t *testing.T) {
	methodeArticleMsg := buildMethodeArticleDeleteMsg()
	mapper := &mapper{}

	_, _, err := mapper.mapMessage(methodeArticleMsg)
	assert.EqualError(t, err, "Methode content is not a content placeholder", "The mapping of the article should be unsuccessful")
}

func TestDoNotHandleMethodeArticleDeleteEvent(t *testing.T) {
	producerMock := new(QueueProducerMock)

	mapper := &mapper{messageProducer: producerMock}

	methodeArticleMsg := buildMethodeArticleDeleteMsg()
	mapper.HandlePlaceholderMessages(methodeArticleMsg)

	producerMock.AssertNotCalled(t, "SendMessage")
}

func TestNotHandleMethodePlaceholderEventWhenProducerReturnsError(t *testing.T) {
	producerMock := new(QueueProducerMock)
	producerMock.On("SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message")).Return(errors.New("I do not want to send the message! I'm on strike!"))

	mapper := &mapper{messageProducer: producerMock}

	methodeMsg := buildIgMethodePlaceholderUpdateMsg()
	mapper.HandlePlaceholderMessages(methodeMsg)

	producerMock.AssertCalled(t, "SendMessage", mock.AnythingOfType("string"), mock.AnythingOfType("producer.Message"))
}

func buildMethodeArticleDeleteMsg() consumer.Message {
	return buildMethodeMsg("test_resources/methode_article_delete.json")
}

func TestDoNotMapPlaceholderWithNoURLInHeadline(t *testing.T) {
	placeholderMsg := buildIgMethodePlaceholderNoURLUpdateMsg()
	mapper := &mapper{}

	_, _, err := mapper.mapMessage(placeholderMsg)
	assert.EqualError(t, err, "Methode Content headline does not contain a link", "The mapping of the placeholder should be unsuccessful")
}

func buildIgMethodePlaceholderNoURLUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_no_url.json")
}

func TestDoNotMapPlaceholderWithWrongURLInHeadline(t *testing.T) {
	placeholderMsg := buildIgMethodePlaceholderWithWrongURLUpdateMsg()
	mapper := &mapper{}

	_, _, err := mapper.mapMessage(placeholderMsg)
	assert.EqualError(t, err, "Methode Content headline does not contain a valid URL - parse pippo: invalid URI for request", "The mapping of the placeholder should be unsuccessful")
}

func buildIgMethodePlaceholderWithWrongURLUpdateMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder_wrong_url.json")
}

func TestDoNotHandleBrightcoveVideoEvent(t *testing.T) {
	producerMock := new(QueueProducerMock)

	mapper := &mapper{messageProducer: producerMock}

	videoMsg := buildBrightcoveVideoMsg()
	mapper.HandlePlaceholderMessages(videoMsg)

	producerMock.AssertNotCalled(t, "SendMessage")
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
	mapper := &mapper{}

	_, _, err := mapper.mapMessage(badPlaceholderMsg)

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
			"Origin-System-Id":  methodeSystemID,
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

func TestSuccesfulBuildOfPlaceholderFromHTTPRequest(t *testing.T) {
	placeholderBody, err := ioutil.ReadFile("test_resources/ig_methode_placeholder_update.json")
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(placeholderBody))
	mapper := New()
	methodePlacheholder, err := mapper.NewMethodeContentPlaceholderFromHTTPRequest(req)
	assert.Nil(t, err, "It should not return an error")
	assert.NotZero(t, methodePlacheholder, "pippo")
}

func TestUnSuccesfulBuildOfPlaceholderFromHTTPRequest(t *testing.T) {
	methodeArticleBody, err := ioutil.ReadFile("test_resources/methode_article_delete.json")
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewReader(methodeArticleBody))
	mapper := New()
	_, err = mapper.NewMethodeContentPlaceholderFromHTTPRequest(req)
	assert.EqualError(t, err, "Methode content is not a content placeholder", "It should return an error")
}
