package message

import (
	"io/ioutil"
	"regexp"
	"testing"
	"time"

	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
)

const expectedTransactionID = "tid_bh7VTFj9Il"

var uuidRegexp = regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")

func TestCorrectMessageToUpdateEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderUpdateMsg()
	expectedCPHPubEventMsg := buildIgPlaceholderPubEvent()
	expectedCCPubEventMsg := buildIgComplementaryContentPubEvent()

	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := NewDefaultCPHMessageCreator()

	methodePlaceholder, _ := model.NewMethodeContentPlaceholder([]byte(igMethodePlaceHolderMsg.Body), igMethodePlaceHolderMsg.Headers["X-Request-Id"], igMethodePlaceHolderMsg.Headers["Message-Timestamp"])
	transformedContents, err := aggregateMapper.MapContentPlaceholder(methodePlaceholder)

	actualCPHPubEventMsg, err := messageCreator.ToPublicationEventMessage(transformedContents[0].GetUppCoreContent(), transformedContents[0])
	assert.Nil(t, err, "It should not return error in creating cph content message")

	actualCCPubEventMsg, err := messageCreator.ToPublicationEventMessage(transformedContents[1].GetUppCoreContent(), transformedContents[1])
	assert.Nil(t, err, "It should not return error in creating complementary content message")

	verifyMessageIsCorrect(t, actualCPHPubEventMsg, &expectedCPHPubEventMsg)
	verifyMessageIsCorrect(t, actualCCPubEventMsg, &expectedCCPubEventMsg)
}

func verifyMessageIsCorrect(t *testing.T, actualPubEventMsg *producer.Message, expectedPubEventMsg *producer.Message) {
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
	return buildMethodeMsg("../mapper/test_resources/ig_methode_placeholder_update.json")
}

func buildIgPlaceholderPubEvent() producer.Message {
	return buildProducerMessage("../mapper/test_resources/ig_placeholder_pub_event.json")
}

func buildIgComplementaryContentPubEvent() producer.Message {
	return buildProducerMessage("../mapper/test_resources/ig_complementarycontent_pub_event.json")
}

func TestCorrectMessageToDeleteEvent(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderDeleteMsg()
	expectedCPHPubEventMsg := buildIgPlaceholderDeleteEvent()
	expectedCCPubEventMsg := buildIgComplementaryContentDeleteEvent()

	aggregateMapper := mapper.NewAggregateCPHMapper()
	messageCreator := NewDefaultCPHMessageCreator()

	methodePlaceholder, _ := model.NewMethodeContentPlaceholder([]byte(igMethodePlaceHolderMsg.Body), igMethodePlaceHolderMsg.Headers["X-Request-Id"], igMethodePlaceHolderMsg.Headers["Message-Timestamp"])
	transformedContents, err := aggregateMapper.MapContentPlaceholder(methodePlaceholder)

	actualCPHPubEventMsg, err := messageCreator.ToPublicationEventMessage(transformedContents[0].GetUppCoreContent(), transformedContents[0])
	assert.Nil(t, err, "It should not return error in creating cph content message")

	actualCCPubEventMsg, err := messageCreator.ToPublicationEventMessage(transformedContents[1].GetUppCoreContent(), transformedContents[1])
	assert.Nil(t, err, "It should not return error in creating complementary content message")

	verifyMessageIsCorrect(t, actualCPHPubEventMsg, &expectedCPHPubEventMsg)
	verifyMessageIsCorrect(t, actualCCPubEventMsg, &expectedCCPubEventMsg)
}

func buildIgMethodePlaceholderDeleteMsg() consumer.Message {
	return buildMethodeMsg("../mapper/test_resources/ig_methode_placeholder_delete.json")
}

func buildIgPlaceholderDeleteEvent() producer.Message {
	return buildProducerMessage("../mapper/test_resources/ig_placeholder_delete_event.json")
}

func buildIgComplementaryContentDeleteEvent() producer.Message {
	return buildProducerMessage("../mapper/test_resources/ig_complementarycontent_delete_event.json")
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

func jsonStringToMap(marshalled string, t *testing.T) map[string]interface{} {
	var unmarshalled map[string]interface{}
	err := json.Unmarshal([]byte(marshalled), &unmarshalled)
	assert.NoError(t, err, "Unmashalling the json content has encountered and error")
	return unmarshalled
}
