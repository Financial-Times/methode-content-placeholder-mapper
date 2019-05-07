package mapper

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestExternalPlaceholderComplementary_Ok(t *testing.T) {
	mockClient := new(model.MockDocStoreClient)
	ccMapper := NewComplementaryContentCPHMapper("api.ft.com", mockClient)

	uppContents, err := ccMapper.MapContentPlaceholder(getPlaceholder(), "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.NoError(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, "lead headline", uppContents[0].(*model.UppComplementaryContent).AlternativeTitles.PromotionalTitle)
	assert.Equal(t, "http://api.ft.com/content/abffff60-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppComplementaryContent).AlternativeImages.PromotionalImage.Id)
	assert.Equal(t, "long standfirst", uppContents[0].(*model.UppComplementaryContent).AlternativeStandfirsts.PromotionalStandfirst)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestExternalPlaceholderComplementaryDelete_Ok(t *testing.T) {
	mockClient := new(model.MockDocStoreClient)
	ccMapper := NewComplementaryContentCPHMapper("api.ft.com", mockClient)

	uppContents, err := ccMapper.MapContentPlaceholder(getDeletedPlaceholder(), "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.NoError(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, true, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestInternalPlaceholderComplementary_Ok(t *testing.T) {
	mockClient := new(model.MockDocStoreClient)
	mockClient.On("GetContent", "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il").Return(getDocStoreContent(t, "document_store_content.json"), nil)
	ccMapper := NewComplementaryContentCPHMapper("api.ft.com", mockClient)

	uppContents, err := ccMapper.MapContentPlaceholder(getPlaceholder(), "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.NoError(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, "lead headline", uppContents[0].(*model.UppComplementaryContent).AlternativeTitles.PromotionalTitle)
	assert.Equal(t, "http://api.ft.com/content/abffff60-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppComplementaryContent).AlternativeImages.PromotionalImage.Id)
	assert.Equal(t, "long standfirst", uppContents[0].(*model.UppComplementaryContent).AlternativeStandfirsts.PromotionalStandfirst)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestInternalPlaceholderComplementaryDelete_Ok(t *testing.T) {
	mockClient := new(model.MockDocStoreClient)
	mockClient.On("GetContent", "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il").Return(getDocStoreContent(t, "document_store_content.json"), nil)
	ccMapper := NewComplementaryContentCPHMapper("api.ft.com", mockClient)

	uppContents, err := ccMapper.MapContentPlaceholder(getDeletedPlaceholder(), "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.NoError(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, false, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestInternalPlaceholderComplementary_DocumentStoreClientError(t *testing.T) {
	mockClient := new(model.MockDocStoreClient)
	mockClient.On("GetContent", "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il").Return(&model.DocStoreUppContent{}, errors.New("DocStore error"))
	ccMapper := NewComplementaryContentCPHMapper("api.ft.com", mockClient)

	_, err := ccMapper.MapContentPlaceholder(getPlaceholder(), "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")
	assert.Error(t, err)
}

func getPlaceholder() *model.MethodeContentPlaceholder {
	return &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           false,
		},
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "lead headline",
			},
			LeadImage: model.LeadImage{
				FileRef: "FT/images/img.jpg?uuid=abffff60-d41a-4a56-8eca-d0f8f0fac068",
			},
			LongStandfirst: "long standfirst",
		},
	}
}

func getDeletedPlaceholder() *model.MethodeContentPlaceholder {
	return &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           true,
		},
	}
}

func getDocStoreContent(t *testing.T, filename string) *model.DocStoreUppContent {
	contentAsBytes, err := ioutil.ReadFile("test_resources/" + filename)
	assert.NoError(t, err, "Unable to open test resource file")
	var docStoreContent model.DocStoreUppContent
	err = json.Unmarshal(contentAsBytes, &docStoreContent)
	assert.NoError(t, err, "Unable to unmarshal test resource file content")
	return &docStoreContent
}
