package mapper

import (
	"testing"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestInternalPlaceholder_Ok(t *testing.T) {
	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           false,
		},
	}
	contentMapper := &ContentCPHMapper{}

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
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

	uppContents, err := contentMapper.MapContentPlaceholder(placeholder, "")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.True(t, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
}
