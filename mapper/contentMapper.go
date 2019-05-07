package mapper

import (
	"fmt"
	"strings"
	"time"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

const (
	placeholderContentURI   = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"
	methodeAuthority        = "http://api.ft.com/system/FTCOM-METHODE"
	verify                  = "verify"
	contentType             = "Content"
	methodeDateFormat       = "20060102150405"
	canonicalWebUrlTemplate = "https://www.ft.com/content/%s"
	webUrlTemplate          = "https://www.ft.com/content/%s"
)

type ContentCPHMapper struct {
}

func (cm *ContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder, uuid, tid, lmd string) ([]model.UppContent, error) {
	if uuid != "" {
		return []model.UppContent{}, nil
	}

	if mcp.Attributes.IsDeleted {
		return []model.UppContent{cm.mapToUppContentPlaceholderDelete(mcp, tid, lmd)}, nil
	}

	uppContent, err := cm.mapToUppContentPlaceholder(mcp, tid, lmd)
	if err != nil {
		return nil, err
	}
	return []model.UppContent{uppContent}, nil
}

func (cm *ContentCPHMapper) mapToUppContentPlaceholder(mpc *model.MethodeContentPlaceholder, tid, lmd string) (*model.UppContentPlaceholder, error) {
	publishDate, err := buildPublishedDate(mpc.Attributes.LastPublicationDate)
	if err != nil {
		return nil, err
	}

	webUrl := mpc.Body.LeadHeadline.URL
	if webUrl == "" {
		webUrl = fmt.Sprintf(webUrlTemplate, mpc.UUID)
	}

	return &model.UppContentPlaceholder{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: tid,
			LastModified:     lmd,
			ContentURI:       placeholderContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted},
		PublishedDate:     publishDate,
		Title:             mpc.Body.LeadHeadline.Text,
		Identifiers:       buildIdentifiers(mpc.UUID),
		WebURL:            webUrl,
		CanonicalWebUrl:   fmt.Sprintf(canonicalWebUrlTemplate, mpc.UUID),
		AlternativeTitles: buildAlternativeTitles(mpc.Body.ContentPackageHeadline),
		Type:              contentType,
		CanBeSyndicated:   verify,
		CanBeDistributed:  verify,
	}, nil
}

func (cm *ContentCPHMapper) mapToUppContentPlaceholderDelete(mpc *model.MethodeContentPlaceholder, tid, lmd string) *model.UppContentPlaceholder {
	return &model.UppContentPlaceholder{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: tid,
			LastModified:     lmd,
			ContentURI:       placeholderContentURI,
			IsMarkedDeleted:  true,
		},
		Type: contentType,
	}
}

func buildIdentifiers(uuid string) []model.Identifier {
	id := model.Identifier{
		Authority:       methodeAuthority,
		IdentifierValue: uuid,
	}
	return []model.Identifier{id}
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
