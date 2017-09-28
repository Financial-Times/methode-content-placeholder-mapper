package mapper

import (
	"testing"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

const tid = "tid_bh7VTFj9Il"

func TestExternalPlaceholder_Ok(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           false,
		},
	}
	contentMapper := &ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "", tid)

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
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

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "", tid)

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

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", tid)

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

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", tid)

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

	_, err := contentMapper.MapContentPlaceholder(placeholder, "", tid)

	assert.NotNil(t, err, "Error was expected during MapContentPlaceholder")
}