package model

import (
	"encoding/xml"
)

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
}

// Attributes is the data structure that models methode content placeholders attributes
type Attributes struct {
	XMLName             xml.Name `xml:"ObjectMetadata"`
	SourceCode          string   `xml:"EditorialNotes>Sources>Source>SourceCode"`
	GenericRefID        string   `xml:"EditorialNotes>Sources>Source>RefId"`
	LastPublicationDate string   `xml:"OutputChannels>DIFTcom>DIFTcomLastPublication"`
	RefField            string   `xml:"WiresIndexing>ref_field"`
	ServiceId           string   `xml:"WiresIndexing>serviceid"`
	Category            string   `xml:"WiresIndexing>category"`
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
