package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Sirupsen/logrus"
)

var blogCategories = []string{"blog", "webchat-live-blogs", "webchat-live-qa", "webchat-markets-live", "fastft"}

type CPHMapper interface {
	MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, uuid string) ([]model.UppContent, *utility.MappingError)
}

type AggregateCPHMapper struct {
	iResolver    IResolver
	cphMappers   []CPHMapper
	cphValidator CPHValidator
}

func NewAggregateCPHMapper(iResolver IResolver, validator CPHValidator, cphMappers []CPHMapper) *AggregateCPHMapper {
	return &AggregateCPHMapper{iResolver: iResolver, cphValidator: validator, cphMappers: cphMappers}
}

func (m *AggregateCPHMapper) MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, uuid string) ([]model.UppContent, *utility.MappingError) {
	err := m.cphValidator.Validate(mpc)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	if m.isBlogCategory(mpc) {
		resolvedUuid, err := m.iResolver.ResolveIdentifier(mpc.Attributes.ServiceId, mpc.Attributes.RefField)
		if err != nil {
			logrus.Warnf("Couldn't resolve blog uuid %v", err)
		} else {
			uuid = resolvedUuid
		}
	}

	var transformedResults []model.UppContent
	for _, cphMapper := range m.cphMappers {
		transformedContents, err := cphMapper.MapContentPlaceholder(mpc, uuid)
		if err != nil {
			return nil, err
		}
		for _, transformedContent := range transformedContents {
			transformedResults = append(transformedResults, transformedContent)
		}
	}
	return transformedResults, nil
}

func (m *AggregateCPHMapper) isBlogCategory(mcp *model.MethodeContentPlaceholder) bool {
	for _, c := range blogCategories {
		if c == mcp.Attributes.Category {
			return true
		}
	}
	return false
}
