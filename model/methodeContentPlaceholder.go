package model

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

const contentPlaceholderSourceCode = "ContentPlaceholder"
const eomCompoundStory = "EOM::CompoundStory"

// MethodeContentPlaceholder is a data structure that models native methode content placeholders
type MethodeContentPlaceholder struct {
	AttributesXML    string      `json:"attributes"`
	SystemAttributes string      `json:"systemAttributes"`
	Type             string      `json:"type"`
	UsageTickets     string      `json:"usageTickets"`
	UUID             string      `json:"uuid"`
	Value            string      `json:"value"`
	WorkflowStatus   string      `json:"workflowStatus"`
	Attributes       Attributes  `json:"-"`
	Body             MethodeBody `json:"-"`
	TransactionID    string      `json:"-"`
	LastModified     string      `json:"-"`
}

// Attributes is the data structure that models methode content placeholders attributes
type Attributes struct {
	XMLName             xml.Name `xml:"ObjectMetadata"`
	SourceCode          string   `xml:"EditorialNotes>Sources>Source>SourceCode"`
	LastPublicationDate string   `xml:"OutputChannels>DIFTcom>DIFTcomLastPublication"`
	LinkedArticleUUID   string   `xml:"WiresIndexing>ref_field"`
	IsDeleted           bool     `xml:"OutputChannels>DIFTcom>DIFTcomMarkDeleted"`
}

// MethodeBody represents the body of a methode content placeholder
type MethodeBody struct {
	XMLName                xml.Name     `xml:"doc"`
	LeadHeadline           LeadHeadline `xml:"lead>lead-headline>headline>ln>a"`
	LeadImage              LeadImage    `xml:"lead>lead-images>web-master"`
	LongStandfirst         string       `xml:"lead>web-stand-first>p"`
	ContentPackageHeadline string       `xml:"lead>package-navigation-headline>ln"`
}

// LeadHeadline represents the LeadHeadline of a content placeholder
type LeadHeadline struct {
	Text string `xml:",chardata"`
	URL  string `xml:"href,attr"`
}

// LeadImage represents the image attribute of a methode content placeholder
type LeadImage struct {
	FileRef string `xml:"fileref,attr"`
}

func (mcp MethodeContentPlaceholder) BuildAttributes(attributesXML string) (Attributes, error) {
	var attrs Attributes
	if err := xml.Unmarshal([]byte(attributesXML), &attrs); err != nil {
		return Attributes{}, err
	}
	return attrs, nil
}

func (mcp MethodeContentPlaceholder) BuildMethodeBody(methodeBodyXMLBase64 string) (MethodeBody, error) {
	methodeBodyXML, err := base64.StdEncoding.DecodeString(methodeBodyXMLBase64)
	if err != nil {
		return MethodeBody{}, err
	}
	var body MethodeBody
	if err := xml.Unmarshal([]byte(methodeBodyXML), &body); err != nil {
		return MethodeBody{}, err
	}
	return body, nil
}

func (mcp MethodeContentPlaceholder) IsInternalCPH() bool {
	return mcp.Attributes.LinkedArticleUUID != ""
}

func NewMethodeContentPlaceholder(messageBody []byte, transactionID string, lastModified string) (*MethodeContentPlaceholder, *utility.MappingError) {
	var p MethodeContentPlaceholder
	if err := json.Unmarshal(messageBody, &p); err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error())
	}
	if p.Type != eomCompoundStory {
		return nil, utility.NewMappingError().WithMessage("Methode content has not type " + eomCompoundStory).ForContent(p.UUID)
	}

	p.TransactionID = transactionID
	p.LastModified = lastModified

	attrs, err := p.BuildAttributes(p.AttributesXML)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.Attributes = attrs

	if p.Attributes.SourceCode != contentPlaceholderSourceCode {
		return nil, utility.NewMappingError().WithMessage("Methode content is not a content placeholder").ForContent(p.UUID)
	}

	body, err := p.BuildMethodeBody(p.Value)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.Body = body

	return &p, nil
}
