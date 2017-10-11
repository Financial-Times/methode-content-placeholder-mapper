package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExternalPlaceholderComplementary_Ok(t *testing.T) {
	ccMapper := ComplementaryContentCPHMapper{}

	placeholder := &model.MethodeContentPlaceholder{
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

	uppContents, err := ccMapper.MapContentPlaceholder(placeholder, "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, "lead headline", uppContents[0].(*model.UppComplementaryContent).AlternativeTitles.PromotionalTitle)
	assert.Equal(t, "abffff60-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppComplementaryContent).AlternativeImages.PromotionalImage)
	assert.Equal(t, "long standfirst", uppContents[0].(*model.UppComplementaryContent).AlternativeStandfirsts.PromotionalStandfirst)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, []model.Brand{{ID: ftBrand}}, uppContents[0].(*model.UppComplementaryContent).Brands)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestExternalPlaceholderComplementaryUpdate_Ok(t *testing.T) {
	ccMapper := ComplementaryContentCPHMapper{}

	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           true,
		},
	}

	uppContents, err := ccMapper.MapContentPlaceholder(placeholder, "", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "e1f02660-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, true, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
}

func TestInternalPlaceholderComplementary_Ok(t *testing.T) {
	ccMapper := ComplementaryContentCPHMapper{}

	placeholder := &model.MethodeContentPlaceholder{
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

	uppContents, err := ccMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents), "Should be one")
	assert.Equal(t, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, "lead headline", uppContents[0].(*model.UppComplementaryContent).AlternativeTitles.PromotionalTitle)
	assert.Equal(t, "abffff60-d41a-4a56-8eca-d0f8f0fac068", uppContents[0].(*model.UppComplementaryContent).AlternativeImages.PromotionalImage)
	assert.Equal(t, "long standfirst", uppContents[0].(*model.UppComplementaryContent).AlternativeStandfirsts.PromotionalStandfirst)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
	assert.Equal(t, []model.Brand{{ID: ftBrand}}, uppContents[0].(*model.UppComplementaryContent).Brands)
	assert.Equal(t, "Content", uppContents[0].(*model.UppComplementaryContent).Type)
}

func TestInternalPlaceholderComplementaryDelete_Ok(t *testing.T) {
	ccMapper := ComplementaryContentCPHMapper{}

	placeholder := &model.MethodeContentPlaceholder{
		UUID: "e1f02660-d41a-4a56-8eca-d0f8f0fac068",
		Attributes: model.Attributes{
			LastPublicationDate: "20140805134048",
			RefField:            "",
			IsDeleted:           true,
		},
	}

	uppContents, err := ccMapper.MapContentPlaceholder(placeholder, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il", "2017-09-27T15:00:00.000Z")

	assert.Nil(t, err, "Error wasn't expected during MapContentPlaceholder")
	assert.Equal(t, 1, len(uppContents))
	assert.Equal(t, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", uppContents[0].GetUUID())
	assert.Equal(t, false, uppContents[0].GetUppCoreContent().IsMarkedDeleted)
	assert.Equal(t, "2017-09-27T15:00:00.000Z", uppContents[0].GetUppCoreContent().LastModified)
	assert.Equal(t, "tid_bh7VTFj9Il", uppContents[0].GetUppCoreContent().PublishReference)
}
