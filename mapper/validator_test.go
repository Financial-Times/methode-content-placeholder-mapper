package mapper

import (
	"testing"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func TestValidatorHeadline_Ok(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "some lead headline",
				URL:  "https://www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.NoError(t, err, "No error should be thrown for valid lead headline.")
}

func TestValidatorHeadlineMissingText_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				URL: "https://www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for missing text in lead headline.")
}

func TestValidatorHeadlineEmptyText_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "",
				URL:  "https://www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for empty text in lead headline.")
}

func TestValidatorHeadlineMissingURL_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "some lead headline",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for missing URL in lead headline.")
}

func TestValidatorHeadlineEmptyURL_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "some lead headline",
				URL:  "",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for empty URL in lead headline.")
}

func TestValidatorHeadlineRelativeURL_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "some lead headline",
				URL:  "www.ft.com/content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for relative URL in lead headline.")
}

func TestValidatorHeadlineInvalidURL_ThrowsError(t *testing.T) {
	defaultValidator := defaultCPHValidator{}

	mcp := &model.MethodeContentPlaceholder{
		Body: model.MethodeBody{
			LeadHeadline: model.LeadHeadline{
				Text: "some lead headline",
				URL:  "content/e1f02660-d41a-4a56-8eca-d0f8f0fac068",
			},
		},
	}

	err := defaultValidator.Validate(mcp)

	assert.Error(t, err, "Error should be thrown for invalid URL in lead headline.")
}
