package mapper

import (
	"fmt"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

var blogCategories = []string{"blog", "webchat-live-blogs", "webchat-live-qa", "webchat-markets-live", "fastft"}

type CPHAggregateMapper interface {
	MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, tid, lmd string) ([]model.UppContent, error)
}

type CPHMapper interface {
	MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, uuid, tid, lmd string) ([]model.UppContent, error)
}

type DefaultCPHAggregateMapper struct {
	iResolver    IResolver
	cphMappers   []CPHMapper
	cphValidator CPHValidator
}

func NewAggregateCPHMapper(iResolver IResolver, validator CPHValidator, cphMappers []CPHMapper) *DefaultCPHAggregateMapper {
	return &DefaultCPHAggregateMapper{iResolver: iResolver, cphValidator: validator, cphMappers: cphMappers}
}

func (m *DefaultCPHAggregateMapper) MapContentPlaceholder(mpc *model.MethodeContentPlaceholder, tid, lmd string) ([]model.UppContent, error) {
	err := m.cphValidator.Validate(mpc)
	if err != nil {
		return nil, err
	}

	uuid := ""
	if m.isBlogCategory(mpc) {
		resolvedUuid, err := m.iResolver.ResolveIdentifier(mpc.Attributes.ServiceId, mpc.Attributes.RefField, tid)
		if err != nil {
			return nil, fmt.Errorf("Couldn't resolve blog uuid %v", err)
		}
		uuid = resolvedUuid
	}

	var transformedResults []model.UppContent
	for _, cphMapper := range m.cphMappers {
		transformedContents, err := cphMapper.MapContentPlaceholder(mpc, uuid, tid, lmd)
		if err != nil {
			return nil, err
		}
		for _, transformedContent := range transformedContents {
			transformedResults = append(transformedResults, transformedContent)
		}
	}
	return transformedResults, nil
}

func (m *DefaultCPHAggregateMapper) isBlogCategory(mcp *model.MethodeContentPlaceholder) bool {
	for _, c := range blogCategories {
		if c == mcp.Attributes.Category {
			return true
		}
	}
	return false
}
