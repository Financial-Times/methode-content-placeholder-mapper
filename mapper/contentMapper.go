package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"strings"
	"time"
)

const (
	placeholderContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"
    methodeAuthority = "http://api.ft.com/system/FTCOM-METHODE"
    verify = "verify"
    contentType = "Content"
	methodeDateFormat = "20060102150405"
	ftBrand = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
)

type ContentCPHMapper struct {
}

func (cm *ContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder, uuid string, tid string) ([]model.UppContent, *utility.MappingError) {
	if uuid != "" {
		return []model.UppContent{}, nil
	}

	if mcp.Attributes.IsDeleted {
		return []model.UppContent{cm.mapToUppContentPlaceholderDelete(mcp)}, nil
	}

	uppContent, err := cm.mapToUppContentPlaceholder(mcp)
	if err != nil {
		return nil, err
	}
	return []model.UppContent{uppContent}, nil
}

func (cm *ContentCPHMapper) mapToUppContentPlaceholder(mpc *model.MethodeContentPlaceholder) (*model.UppContentPlaceholder, *utility.MappingError) {
	publishDate, err := buildPublishedDate(mpc.Attributes.LastPublicationDate)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	return &model.UppContentPlaceholder{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       placeholderContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted},
		PublishedDate:     publishDate,
		Title:             mpc.Body.LeadHeadline.Text,
		Identifiers:       buildIdentifiers(mpc.UUID),
		Brands:            buildBrands(),
		WebURL:            mpc.Body.LeadHeadline.URL,
		AlternativeTitles: buildAlternativeTitles(mpc.Body.ContentPackageHeadline),
		Type:              contentType,
		CanBeSyndicated:   verify,
		CanBeDistributed:  verify,
	}, nil
}

func (cm *ContentCPHMapper) mapToUppContentPlaceholderDelete(mpc *model.MethodeContentPlaceholder) *model.UppContentPlaceholder {
	return &model.UppContentPlaceholder{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       placeholderContentURI,
			IsMarkedDeleted:  true},
	}
}

func buildIdentifiers(uuid string) []model.Identifier {
	id := model.Identifier{
		Authority:       methodeAuthority,
		IdentifierValue: uuid,
	}
	return []model.Identifier{id}
}

func buildBrands() []model.Brand {
	brand := model.Brand{ID: ftBrand}
	return []model.Brand{brand}
}

func buildAlternativeTitles(contentPackageTitle string) *model.AlternativeTitles {
	contentPackageTitle = strings.TrimSpace(contentPackageTitle)

	if contentPackageTitle == "" {
		return nil
	}
	return &model.AlternativeTitles{ContentPackageTitle: contentPackageTitle}
}

func buildPublishedDate(lastPublicationDate string) (string, error) {
	date, err := time.Parse(methodeDateFormat, lastPublicationDate)
	if err != nil {
		return "", err
	}
	return date.Format(model.UPPDateFormat), nil
}
