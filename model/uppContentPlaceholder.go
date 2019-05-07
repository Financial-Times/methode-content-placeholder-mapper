package model

const (
	MethodeSystemID = "http://cmdb.ft.com/systems/methode-web-pub"
	UPPDateFormat   = "2006-01-02T15:04:05.000Z0700"
	ftBrand         = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
)

// UppContentPlaceholder represents the content placeholder representation according to UPP model.
// Note: Title holds the text of alternativeTitle as a CPH does not have a title and some clients expect one.
type UppContentPlaceholder struct {
	UppCoreContent
	PublishedDate     string             `json:"publishedDate"`
	Title             string             `json:"title"`
	Identifiers       []Identifier       `json:"identifiers"`
	AlternativeTitles *AlternativeTitles `json:"alternativeTitles"`
	WebURL            string             `json:"webUrl"`
	CanonicalWebUrl   string             `json:"canonicalWebUrl"`
	Type              string             `json:"type"`
	CanBeSyndicated   string             `json:"canBeSyndicated"`
	CanBeDistributed  string             `json:"canBeDistributed"`
}

// Identifier represents content identifiers according to UP data model
type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

// AlternativeTitles represents the alternative titles for content according to UP data model
type AlternativeTitles struct {
	PromotionalTitle    string `json:"promotionalTitle,omitempty"`
	ContentPackageTitle string `json:"contentPackageTitle,omitempty"`
}
