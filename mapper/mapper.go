package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

type CPHMapper interface {
	MapContentPlaceholder(mpc *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError)
}

type AggregateCPHMapper struct {
	cphMappers   []CPHMapper
	cphValidator CPHValidator
}

type contentCPHMapper struct {
}

type complementaryContentCPHMapper struct {
}

func NewAggregateCPHMapper() *AggregateCPHMapper {
	return &AggregateCPHMapper{cphValidator: NewDefaultCPHValidator(), cphMappers: []CPHMapper{&contentCPHMapper{}, &complementaryContentCPHMapper{}}}
}

func (m *AggregateCPHMapper) MapContentPlaceholder(mpc *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError) {
	err := m.cphValidator.Validate(mpc)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	var transformedResults []model.UppContent
	for _, cphMapper := range m.cphMappers {
		transformedContents, err := cphMapper.MapContentPlaceholder(mpc)
		if err != nil {
			return nil, err
		}
		for _, transformedContent := range transformedContents {
			transformedResults = append(transformedResults, transformedContent)
		}
	}
	return transformedResults, nil
}

func (cm *contentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError) {
	if mcp.IsInternalCPH() {
		return []model.UppContent{}, nil
	} else {
		if mcp.Attributes.IsDeleted {
			return []model.UppContent{model.NewUppContentPlaceholderDelete(mcp)}, nil
		}

		uppContent, err := model.NewUppContentPlaceholder(mcp)
		if err != nil {
			return nil, err
		}

		return []model.UppContent{uppContent}, nil
	}
}

func (ccm *complementaryContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError) {
	uuidToSet := mcp.UUID
	if mcp.IsInternalCPH() {
		uuidToSet = mcp.Attributes.LinkedArticleUUID
	}

	if mcp.Attributes.IsDeleted {
		return []model.UppContent{model.NewUppComplementaryContentDelete(mcp, uuidToSet)}, nil
	}

	return []model.UppContent{model.NewUppComplementaryContent(mcp, uuidToSet)}, nil
}
