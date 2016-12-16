package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
)

const contentPlaceholderSourceCode = "ContentPlaceholder"
const eomCompandStory = "EOM::CompoundStory"

type MethodeContentPlaceholder struct {
	AttributesXML    string `json:"attributes"`
	SystemAttributes string `json:"systemAttributes"`
	Type             string `json:"type"`
	UsageTickets     string `json:"usageTickets"`
	UUID             string `json:"uuid"`
	Value            string `json:"value"`
	WorkflowStatus   string `json:"workflowStatus"`
	attributes       Attributes
	body             MethodeBody
	transactionId    string
	lastModified     string
}

type Attributes struct {
	XMLName             xml.Name `xml:"ObjectMetadata"`
	SourceCode          string   `xml:"EditorialNotes>Sources>Source>SourceCode"`
	LastPublicationDate string   `xml:"OutputChannels>DIFTcom>DIFTcomLastPublication"`
}

type MethodeBody struct {
	XMLName        xml.Name     `xml:"doc"`
	LeadHeadline   LeadHeadline `xml:"lead>lead-headline>headline>ln>a"`
	LeadImage      LeadImage    `xml:"lead>lead-images>web-master"`
	LongStandfirst string       `xml:"lead>web-stand-first>p"`
}

type LeadHeadline struct {
	Text string `xml:",chardata"`
	Url  string `xml:"href,attr"`
}

type LeadImage struct {
	FileRef string `xml:"fileref,attr"`
}

func newMethodeContentPlaceholder(msg consumer.Message) (MethodeContentPlaceholder, error) {
	var p MethodeContentPlaceholder
	if err := json.Unmarshal([]byte(msg.Body), &p); err != nil {
		return MethodeContentPlaceholder{}, err
	}
	if p.Type != eomCompandStory {
		return MethodeContentPlaceholder{}, errors.New("Methode content has not type " + eomCompandStory)
	}

	p.transactionId = msg.Headers["X-Request-Id"]
	p.lastModified = msg.Headers["Message-Timestamp"]

	attrs, err := buildAttributes(p.AttributesXML)
	if err != nil {
		return MethodeContentPlaceholder{}, err
	}
	p.attributes = attrs

	body, err := buildMethodeBody(p.Value)
	if err != nil {
		return MethodeContentPlaceholder{}, err
	}
	p.body = body
	fmt.Println(p.attributes.LastPublicationDate)
	fmt.Println(p.body.LeadHeadline.Text)
	if p.attributes.SourceCode != contentPlaceholderSourceCode {
		return MethodeContentPlaceholder{}, errors.New("Methode content is not a content placeholder")
	}
	return p, nil
}

func buildAttributes(attributesXml string) (Attributes, error) {
	var attrs Attributes
	if err := xml.Unmarshal([]byte(attributesXml), &attrs); err != nil {
		return Attributes{}, err
	}
	return attrs, nil
}

func buildMethodeBody(methodeBodyXmlBase64 string) (MethodeBody, error) {
	methodeBodyXml, err := base64.StdEncoding.DecodeString(methodeBodyXmlBase64)
	if err != nil {
		return MethodeBody{}, err
	}
	var body MethodeBody
	if err := xml.Unmarshal([]byte(methodeBodyXml), &body); err != nil {
		return MethodeBody{}, err
	}
	return body, nil
}

type UpContentPlaceholder struct {
	UUID                  string                `json:"uuid"`
	Identifiers           []Identifier          `json:"identifiers"`
	Brands                []Brand               `json:"brands"`
	AlternativeTitles     AlternativeTitles     `json:"alternativeTitles"`
	AlternativeImages     AlternativeImages     `json:"alternativeImages"`
	AlternativeStandfirst AlternativeStandfirst `json:"alternativeStandfirst"`
	PublishedDate         string                `json:"publishedDate"`
	PublishReference      string                `json:"publishReference"`
	LastModified          string                `json:"lastModified"`
	WebUrl                string                `json:"webUrl"`
	CanBeSyndicated       string                `json:"canBeSyndicated"`
}

type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

type Brand struct {
	ID string `json:"id"`
}

type AlternativeTitles struct {
	PromotionalTitle string `json:"promotionalTitle"`
}

type AlternativeImages struct {
	PromotionalImage string `json:"promotionalImage"`
}

type AlternativeStandfirst struct {
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}

func (p UpContentPlaceholder) toProducerMessage() (producer.Message, error) {
	//TODO implement
	// headers := map[string]string{
	// 	"X-Request-Id":      tid,
	// 	"Message-Timestamp": time.Now().Format(dateFormat),
	// 	"Message-Id":        uuid.NewV4().String(),
	// 	"Message-Type":      "cms-content-published",
	// 	"Content-Type":      "application/json",
	// 	"Origin-System-Id":  methodeSystemId,
	// }
	//
	// mappedMsg := producer.Message{Headers: headers, Body: string(marshalledEvent)}
	// return mappedMsg, "", nil
	return producer.Message{}, nil
}
