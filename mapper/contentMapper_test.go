package mapper

import (
	"testing"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestExternalPlaceholder_Ok(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           false,
		},
		Body: model.MethodeBody{
			ContentPackageHeadline: "cp headline",
			LeadHeadline: model.LeadHeadline{
				Text: "lead headline",
				URL: "www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}
	contentMapper := ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, ftBrand, uppContents[0].(*model.UppContentPlaceholder).Brands[0].ID)
	assert.Equal(t, methodeAuthority, uppContents[0].(*model.UppContentPlaceholder).Identifiers[0].Authority)
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppContentPlaceholder).Identifiers[0].IdentifierValue)
	assert.Equal(t, "cp headline", uppContents[0].(*model.UppContentPlaceholder).AlternativeTitles.ContentPackageTitle)
	assert.Equal(t, "2014-08-05T13:40:48.000Z", uppContents[0].(*model.UppContentPlaceholder).PublishedDate)
	assert.Equal(t, "Content", uppContents[0].(*model.UppContentPlaceholder).Type)
	assert.Equal(t, "verify", uppContents[0].(*model.UppContentPlaceholder).CanBeDistributed)
	assert.Equal(t, "lead headline", uppContents[0].(*model.UppContentPlaceholder).Title)
	assert.Equal(t, "www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppContentPlaceholder).WebURL)
	assert.NotEmpty(t, uppContents[0].(*model.UppContentPlaceholder).UppCoreContent.LastModified)
	assert.Equal(t, false, uppContents[0].(*model.UppContentPlaceholder).UppCoreContent.IsMarkedDeleted)
	assert.Equal(t, "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/", uppContents[0].(*model.UppContentPlaceholder).UppCoreContent.ContentURI)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].(*model.UppContentPlaceholder).UppCoreContent.PublishReference)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].(*model.UppContentPlaceholder).UppCoreContent.LastModified)
}

func TestExternalPlaceholderDeleted_Ok(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           true,
		},
	}
	contentMapper := ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.True(t, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
}

func TestInternalPlaceholder_ReturnsEmpty(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           false,
		},
	}
	contentMapper := &ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 0, len(uppContents), "Should be zero")
}

func TestInternalPlaceholderDeleted_Ok(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           true,
		},
	}
	contentMapper := ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 0, len(uppContents))
}

func TestExternalPlaceholderWrongPublishDate_ThrowsError(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "201408051340",
			RefField:            "",
			IsDeleted:           false,
		},
	}
	contentMapper := &ContentCPHMapper{}

	_, err := contentMapper.MapContentPlaceholder(placeholder, "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.NotNil(t, err, "Error was expected during MapContentPlaceholder")
}
