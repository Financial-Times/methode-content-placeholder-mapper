package model

import (
	"strings"
	"time"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

const placeholderContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"
const UPPDateFormat = "2006-01-02T15:04:05.000Z0700"

const MethodeSystemID = "http://cmdb.ft.com/systems/methode-web-pub"
const methodeAuthority = "http://api.ft.com/system/FTCOM-METHODE"
const methodeDateFormat = "20060102150405"

const ftBrand = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"

const canBeDistributedVerify = "verify"

// UppContentPlaceholder represents the content placeholder representation according to UPP model.
// Note: Title holds the text of alternativeTitle as a CPH does not have a title and some clients expect one.
type UppContentPlaceholder struct {
	UppCoreContent
	PublishedDate     string             `json:"publishedDate"`
	Title             string             `json:"title"`
	Identifiers       []Identifier       `json:"identifiers"`
	Brands            []Brand            `json:"brands"`
	AlternativeTitles *AlternativeTitles `json:"alternativeTitles"`
	WebURL            string             `json:"webUrl"`
	Type              string             `json:"type"`
	CanBeSyndicated   string             `json:"canBeSyndicated"`
	CanBeDistributed  string             `json:"canBeDistributed"`
}

// Identifier represents content identifiers according to UP data model
type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

// Brand represents a content brand according to UP data model
type Brand struct {
	ID string `json:"id"`
}

// AlternativeTitles represents the alternative titles for content according to UP data model
type AlternativeTitles struct {
	PromotionalTitle    string `json:"promotionalTitle,omitempty"`
	ContentPackageTitle string `json:"contentPackageTitle,omitempty"`
}

func NewUppContentPlaceholder(mpc *MethodeContentPlaceholder) (*UppContentPlaceholder, *utility.MappingError) {
	publishDate, err := buildPublishedDate(mpc.Attributes.LastPublicationDate)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	return &UppContentPlaceholder{
		UppCoreContent: UppCoreContent{
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
		Type:              "Content",
		CanBeSyndicated:   "verify",
		CanBeDistributed:  canBeDistributedVerify,
	}, nil
}

func NewUppContentPlaceholderDelete(mpc *MethodeContentPlaceholder) *UppContentPlaceholder {
	return &UppContentPlaceholder{
		UppCoreContent: UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       placeholderContentURI,
			IsMarkedDeleted:  true},
	}
}

func buildIdentifiers(uuid string) []Identifier {
	id := Identifier{
		Authority:       methodeAuthority,
		IdentifierValue: uuid,
	}
	return []Identifier{id}
}

func buildBrands() []Brand {
	brand := Brand{ID: ftBrand}
	return []Brand{brand}
}

func buildAlternativeTitles(contentPackageTitle string) *AlternativeTitles {
	contentPackageTitle = strings.TrimSpace(contentPackageTitle)

	if contentPackageTitle == "" {
		return nil
	}
	return &AlternativeTitles{ContentPackageTitle: contentPackageTitle}
}

func buildPublishedDate(lastPublicationDate string) (string, error) {
	date, err := time.Parse(methodeDateFormat, lastPublicationDate)
	if err != nil {
		return "", err
	}
	return date.Format(UPPDateFormat), nil
}
