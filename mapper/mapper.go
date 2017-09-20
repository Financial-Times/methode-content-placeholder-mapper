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

func NewAggregateCPHMapper(validator CPHValidator, cphMappers []CPHMapper) *AggregateCPHMapper {
	return &AggregateCPHMapper{cphValidator: validator, cphMappers: cphMappers}
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
