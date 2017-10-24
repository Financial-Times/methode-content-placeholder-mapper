package message

import (
	"encoding/json"
	"testing"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestToPublicationEventUpdate_Ok(t *testing.T) {
	defaultMessageCreator := NewDefaultCPHMessageCreator()

	coreContentCPH := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/",
		IsMarkedDeleted:  false,
	}

	coreContentCompContent := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/",
		IsMarkedDeleted:  false,
	}

	cphContent := model.UppContentPlaceholder{}
	cContent := model.UppComplementaryContent{}

	expectedCPHPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		Payload:      cphContent,
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	expectedCompContentPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		Payload:      cContent,
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	actualCPHPubEvent := defaultMessageCreator.ToPublicationEvent(coreContentCPH, cphContent)
	actualCompContentPubEvent := defaultMessageCreator.ToPublicationEvent(coreContentCompContent, cContent)

	assert.Equal(t, expectedCPHPubEvent.Payload, actualCPHPubEvent.Payload)
	assert.Equal(t, expectedCPHPubEvent.LastModified, actualCPHPubEvent.LastModified)
	assert.Equal(t, expectedCPHPubEvent.ContentURI, actualCPHPubEvent.ContentURI)

	assert.Equal(t, expectedCompContentPubEvent.Payload, actualCompContentPubEvent.Payload)
	assert.Equal(t, expectedCompContentPubEvent.LastModified, actualCompContentPubEvent.LastModified)
	assert.Equal(t, expectedCompContentPubEvent.ContentURI, actualCompContentPubEvent.ContentURI)
}

func TestToPublicationEventDelete_Ok(t *testing.T) {
	defaultMessageCreator := NewDefaultCPHMessageCreator()

	coreContentCPH := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/",
		IsMarkedDeleted:  true,
	}

	coreContentCompContent := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/",
		IsMarkedDeleted:  true,
	}

	cphContent := model.UppContentPlaceholder{}
	cContent := model.UppComplementaryContent{}

	expectedCPHPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	expectedCompContentPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	actualCPHPubEvent := defaultMessageCreator.ToPublicationEvent(coreContentCPH, cphContent)
	actualCompContentPubEvent := defaultMessageCreator.ToPublicationEvent(coreContentCompContent, cContent)

	assert.Nil(t, actualCPHPubEvent.Payload)
	assert.Equal(t, expectedCPHPubEvent.LastModified, actualCPHPubEvent.LastModified)
	assert.Equal(t, expectedCPHPubEvent.ContentURI, actualCPHPubEvent.ContentURI)

	assert.Nil(t, actualCompContentPubEvent.Payload)
	assert.Equal(t, expectedCompContentPubEvent.LastModified, actualCompContentPubEvent.LastModified)
	assert.Equal(t, expectedCompContentPubEvent.ContentURI, actualCompContentPubEvent.ContentURI)
}

func TestToPublicationEventMsgUpdate_Ok(t *testing.T) {
	defaultMessageCreator := NewDefaultCPHMessageCreator()

	coreContentCPH := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/",
		IsMarkedDeleted:  false,
	}

	coreContentCompContent := &model.UppCoreContent{
		UUID:             "512c1f3d-e48c-4618-863c-94bc9d913b9b",
		PublishReference: "tid_test123",
		LastModified:     "2017-05-15T15:54:32.166Z",
		ContentURI:       "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/",
		IsMarkedDeleted:  false,
	}

	cphContent := model.UppContentPlaceholder{UppCoreContent: *coreContentCPH}
	cContent := model.UppComplementaryContent{UppCoreContent: *coreContentCompContent}

	expectedCPHPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		Payload:      cphContent,
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	expectedCompContentPubEvent := &model.PublicationEvent{
		ContentURI:   "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/512c1f3d-e48c-4618-863c-94bc9d913b9b",
		Payload:      cContent,
		LastModified: "2017-05-15T15:54:32.166Z",
	}

	headers := map[string]string{
		"X-Request-Id":     "tid_test123",
		"Message-Type":     "cms-content-published",
		"Content-Type":     "application/json",
		"Origin-System-Id": "http://cmdb.ft.com/systems/methode-web-pub",
	}

	cphPubEventMarshalled, _ := json.Marshal(expectedCPHPubEvent)
	expectedCPHPubEventMsg := &producer.Message{Headers: headers, Body: string(cphPubEventMarshalled)}
	compContentPubEventMarshalled, _ := json.Marshal(expectedCompContentPubEvent)
	expectedCompContentPubEventMsg := &producer.Message{Headers: headers, Body: string(compContentPubEventMarshalled)}

	actualCPHPubEventMsg, err := defaultMessageCreator.ToPublicationEventMessage(coreContentCPH, cphContent)
	assert.NoError(t, err, "No error should be thrown.")

	actualCompContentPubEventMsg, err := defaultMessageCreator.ToPublicationEventMessage(coreContentCompContent, cContent)
	assert.NoError(t, err, "No error should be thrown.")

	verifyMessageIsCorrect(t, expectedCPHPubEventMsg, actualCPHPubEventMsg)
	verifyMessageIsCorrect(t, expectedCompContentPubEventMsg, actualCompContentPubEventMsg)
}

func verifyMessageIsCorrect(t *testing.T, actualPubEventMsg, expectedPubEventMsg *producer.Message) {
	assert.Equal(t, "tid_test123", actualPubEventMsg.Headers["X-Request-Id"], "The Transaction ID should be consistent")
	assert.Equal(t, "cms-content-published", actualPubEventMsg.Headers["Message-Type"], "The Message type should be cms-content-published")
	assert.Equal(t, "application/json", actualPubEventMsg.Headers["Content-Type"], "The Content type should be application/json")

	expectedBodyAsMap := jsonStringToMap(expectedPubEventMsg.Body, t)
	actualBodyAsMap := jsonStringToMap(actualPubEventMsg.Body, t)
	assert.Equal(t, expectedBodyAsMap, actualBodyAsMap, "The placeholder should be mapped properly")
}

func jsonStringToMap(marshalled string, t *testing.T) map[string]interface{} {
	var unmarshalled map[string]interface{}
	err := json.Unmarshal([]byte(marshalled), &unmarshalled)
	assert.NoError(t, err, "Unmashalling the json content has encountered and error")
	return unmarshalled
}
