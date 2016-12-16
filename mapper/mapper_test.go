package mapper

import (
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
)

const expectedcunsumerMsgBody = ""

var expectedUpPlaceholder = UpContentPlaceholder{UUID: "3b1f4644-6e97-11e6-a6d8-a8ffd2ee4b1c"}

func TestCorrectMapping(t *testing.T) {
	igMethodePlaceHolderMsg := buildIgMethodePlaceholderMsg()
	igMethodePlaceHolder, err := newMethodeContentPlaceholder(igMethodePlaceHolderMsg)
	assert.Nil(t, err, "No error in building methode placeholder struct")

	mapper := &mapper{}

	actualUpPlaceholder, err := mapper.mapContentPlaceholder(igMethodePlaceHolder)
	assert.Nil(t, err, "No error in mapping placeholder")
	assert.Equal(t, expectedUpPlaceholder, actualUpPlaceholder, "Place holder mapped properly")
}

func buildIgMethodePlaceholderMsg() consumer.Message {
	return buildMethodeMsg("test_resources/ig_methode_placeholder.json")
}

func buildMethodeMsg(examplePath string) consumer.Message {
	placeholderBody, err := ioutil.ReadFile(examplePath)
	if err != nil {
		panic(err)
	}
	return consumer.Message{
		Body: string(placeholderBody),
		Headers: map[string]string{
			"X-Request-Id":      "tid_i1ktygkniy",
			"Message-Timestamp": "2016-12-16T13:13:51.154Z",
		},
	}
}
