package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	tid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
	uuid "github.com/satori/go.uuid"
)

const methodeSystemID = "http://cmdb.ft.com/systems/methode-web-pub"
const methodeDateFormat = "20060102150405"
const contentPlaceholderSourceCode = "ContentPlaceholder"
const eomCompoundStory = "EOM::CompoundStory"

const upDateFormat = "2006-01-02T03:04:05.000Z0700"
const ftBrand = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
const methodeAuthority = "http://api.ft.com/system/FTCOM-METHODE"
const mapperURIBase = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"

const canBeDistributedVerify = "verify"

// Mapper is a generic interface for content paceholder mapper
type Mapper interface {
	HandlePlaceholderMessages(msg consumer.Message)
	StartMappingMessages(c consumer.MessageConsumer, p producer.MessageProducer)
	NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (MethodeContentPlaceholder, *MappingError)
	MapContentPlaceholder(mpc MethodeContentPlaceholder) (UpContentPlaceholder, *MappingError)
}

type mapper struct {
	messageConsumer consumer.MessageConsumer
	messageProducer producer.MessageProducer
}

// New returns a new Mapper instance
func New() Mapper {
	return &mapper{}
}

func (m *mapper) HandlePlaceholderMessages(msg consumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if msg.Headers["Origin-System-Id"] != methodeSystemID {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}
	placeholderMsg, placeholderUUID, mappingErr := m.mapMessage(msg)
	if mappingErr != nil {
		log.WithField("transaction_id", tid).WithField("uuid", mappingErr.ContentUUID).WithError(mappingErr).Warn("Error in mapping message")
		return
	}
	err := m.messageProducer.SendMessage(placeholderUUID, placeholderMsg)
	if err != nil {
		log.WithField("transaction_id", tid).WithField("uuid", placeholderUUID).WithError(err).Warn("Error sending transformed message to queue")
		return
	}
	log.WithField("transaction_id", tid).WithField("uuid", placeholderUUID).Info("Content mapped and sent to the queue")
}

func (m *mapper) mapMessage(msg consumer.Message) (producer.Message, string, *MappingError) {
	methodePlaceholder, err := m.newMethodeContentPlaceholderFromConsumerMessage(msg)
	if err != nil {
		return producer.Message{}, "", err
	}

	upPlaceholder, err := m.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		return producer.Message{}, "", err
	}

	pubEventMsg, err := upPlaceholder.toPublicationEventMessage()
	if err != nil {
		return producer.Message{}, "", err
	}

	return pubEventMsg, upPlaceholder.UUID, nil
}

func (m *mapper) MapContentPlaceholder(mpc MethodeContentPlaceholder) (UpContentPlaceholder, *MappingError) {
	// When a methode placeholder has been delete, we map the message with an empty body
	if mpc.attributes.IsDeleted {
		return UpContentPlaceholder{
			UUID:             mpc.UUID,
			LastModified:     mpc.lastModified,
			PublishReference: mpc.transactionID,
		}, nil
	}
	err := validateHeadline(mpc.body.LeadHeadline)
	if err != nil {
		return UpContentPlaceholder{}, NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	publishDate, err := buildPublishedDate(mpc.attributes.LastPublicationDate)
	if err != nil {
		return UpContentPlaceholder{}, NewMappingError().WithMessage(err.Error()).ForContent(mpc.UUID)
	}

	upPlaceholder := UpContentPlaceholder{
		UUID:                   mpc.UUID,
		Title:                  mpc.body.LeadHeadline.Text,
		Identifiers:            buildIdentifiers(mpc.UUID),
		Brands:                 buildBrands(),
		WebURL:                 mpc.body.LeadHeadline.URL,
		AlternativeTitles:      buildAlternativeTitles(mpc.body.LeadHeadline.Text),
		AlternativeImages:      buildAlternativeImages(mpc.body.LeadImage.FileRef),
		AlternativeStandfirsts: buildAlternativeStandfirsts(mpc.body.LongStandfirst),
		PublishedDate:          publishDate,
		PublishReference:       mpc.transactionID,
		LastModified:           mpc.lastModified,
		Type:                   "Content",
		CanBeSyndicated:        "verify",
		CanBeDistributed:       canBeDistributedVerify,
	}
	return upPlaceholder, nil
}

func validateHeadline(headline LeadHeadline) error {
	if headline.Text == "" {
		return errors.New("Methode Content headline does not contain text")
	}
	if headline.URL == "" {
		return errors.New("Methode Content headline does not contain a link")
	}
	headlineURL, err := url.Parse(headline.URL)
	if err != nil {
		return errors.New("Methode Content headline does not contain a valid URL - " + err.Error())
	}
	if !headlineURL.IsAbs() {
		return errors.New("Methode Content headline does not contain an absolute URL")
	}
	return nil
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

func buildAlternativeTitles(promoTitle string) *AlternativeTitles {
	promoTitle = strings.TrimSpace(promoTitle)
	if promoTitle == "" {
		return nil
	}
	return &AlternativeTitles{PromotionalTitle: promoTitle}
}

func buildAlternativeImages(fileRef string) *AlternativeImages {
	if fileRef == "" {
		return nil
	}
	imageUUID := extractImageUUID(fileRef)
	return &AlternativeImages{PromotionalImage: imageUUID}
}

func extractImageUUID(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func buildAlternativeStandfirsts(promoStandfirst string) *AlternativeStandfirsts {
	promoStandfirst = strings.TrimSpace(promoStandfirst)
	if promoStandfirst == "" {
		return nil
	}
	return &AlternativeStandfirsts{PromotionalStandfirst: promoStandfirst}
}

func buildPublishedDate(lastPublicationDate string) (string, error) {
	date, err := time.Parse(methodeDateFormat, lastPublicationDate)
	if err != nil {
		return "", err
	}
	return date.Format(upDateFormat), nil
}

func (m *mapper) StartMappingMessages(c consumer.MessageConsumer, p producer.MessageProducer) {
	m.messageConsumer = c
	m.messageProducer = p

	log.Infof("Starting queue consumer...")
	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		m.messageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	m.messageConsumer.Stop()
	consumerWaitGroup.Wait()
}

// MethodeContentPlaceholder is a data structure that models native methode content placeholders
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
	transactionID    string
	lastModified     string
}

// Attributes is the data structure that models methode content placeholders attributes
type Attributes struct {
	XMLName             xml.Name `xml:"ObjectMetadata"`
	SourceCode          string   `xml:"EditorialNotes>Sources>Source>SourceCode"`
	LastPublicationDate string   `xml:"OutputChannels>DIFTcom>DIFTcomLastPublication"`
	IsDeleted           bool     `xml:"OutputChannels>DIFTcom>DIFTcomMarkDeleted"`
}

// MethodeBody represents the body of a methode content placeholder
type MethodeBody struct {
	XMLName        xml.Name     `xml:"doc"`
	LeadHeadline   LeadHeadline `xml:"lead>lead-headline>headline>ln>a"`
	LeadImage      LeadImage    `xml:"lead>lead-images>web-master"`
	LongStandfirst string       `xml:"lead>web-stand-first>p"`
}

// LeadHeadline reppresents the LeadHeadline of a content placeholder
type LeadHeadline struct {
	Text string `xml:",chardata"`
	URL  string `xml:"href,attr"`
}

// LeadImage represents the image attribute of a methode content placeholder
type LeadImage struct {
	FileRef string `xml:"fileref,attr"`
}

func (m *mapper) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (MethodeContentPlaceholder, *MappingError) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	lastModified := time.Now().String()
	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage(err.Error())
	}
	return m.newMethodeContentPlaceholder(messageBody, transactionID, lastModified)
}

func (m *mapper) newMethodeContentPlaceholder(messageBody []byte, transactionID string, lastModified string) (MethodeContentPlaceholder, *MappingError) {
	var p MethodeContentPlaceholder
	if err := json.Unmarshal(messageBody, &p); err != nil {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage(err.Error())
	}
	if p.Type != eomCompoundStory {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage("Methode content has not type " + eomCompoundStory).ForContent(p.UUID)
	}

	p.transactionID = transactionID
	p.lastModified = lastModified

	attrs, err := buildAttributes(p.AttributesXML)
	if err != nil {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.attributes = attrs

	if p.attributes.SourceCode != contentPlaceholderSourceCode {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage("Methode content is not a content placeholder").ForContent(p.UUID)
	}

	body, err := buildMethodeBody(p.Value)
	if err != nil {
		return MethodeContentPlaceholder{}, NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.body = body

	return p, nil
}

func (m *mapper) newMethodeContentPlaceholderFromConsumerMessage(msg consumer.Message) (MethodeContentPlaceholder, *MappingError) {
	transactionID := msg.Headers["X-Request-Id"]
	lastModified := msg.Headers["Message-Timestamp"]

	return m.newMethodeContentPlaceholder([]byte(msg.Body), transactionID, lastModified)
}

func buildAttributes(attributesXML string) (Attributes, error) {
	var attrs Attributes
	if err := xml.Unmarshal([]byte(attributesXML), &attrs); err != nil {
		return Attributes{}, err
	}
	return attrs, nil
}

func buildMethodeBody(methodeBodyXMLBase64 string) (MethodeBody, error) {
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

// UpContentPlaceholder reppresents the content placeholder representation according to UP model
//note Title holds the text of alternativeTitle as a cph does not have a title and some clients expect one.
type UpContentPlaceholder struct {
	UUID                   string                  `json:"uuid"`
	Title                  string                  `json:"title"`
	Identifiers            []Identifier            `json:"identifiers"`
	Brands                 []Brand                 `json:"brands"`
	AlternativeTitles      *AlternativeTitles      `json:"alternativeTitles"`
	AlternativeImages      *AlternativeImages      `json:"alternativeImages"`
	AlternativeStandfirsts *AlternativeStandfirsts `json:"alternativeStandfirsts"`
	PublishedDate          string                  `json:"publishedDate"`
	PublishReference       string                  `json:"publishReference"`
	LastModified           string                  `json:"lastModified"`
	WebURL                 string                  `json:"webUrl"`
	Type                   string                  `json:"type"`
	CanBeSyndicated        string                  `json:"canBeSyndicated"`
	CanBeDistributed       string                  `json:"canBeDistributed"`

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
	PromotionalTitle string `json:"promotionalTitle"`
}

// AlternativeImages represents the alternative images for content according to UP data model
type AlternativeImages struct {
	PromotionalImage string `json:"promotionalImage"`
}

// AlternativeStandfirsts represents the alternative standfirsts for content according to UP data model
type AlternativeStandfirsts struct {
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}

func (p UpContentPlaceholder) toPublicationEventMessage() (producer.Message, *MappingError) {

	publicationEvent := p.toPublicationEvent()

	jsonPublicationEvent, err := json.Marshal(publicationEvent)
	if err != nil {
		return producer.Message{}, NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}

	headers := map[string]string{
		"X-Request-Id":      p.PublishReference,
		"Message-Timestamp": time.Now().Format(upDateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      "cms-content-published",
		"Content-Type":      "application/json",
		"Origin-System-Id":  methodeSystemID,
	}

	producerMsg := producer.Message{Headers: headers, Body: string(jsonPublicationEvent)}
	return producerMsg, nil
}

func (p UpContentPlaceholder) toPublicationEvent() publicationEvent {
	if p.isDeleted() {
		return publicationEvent{
			ContentURI:   mapperURIBase + p.UUID,
			LastModified: p.LastModified,
		}
	}
	return publicationEvent{
		ContentURI:   mapperURIBase + p.UUID,
		Payload:      &p,
		LastModified: p.LastModified,
	}
}

func (p UpContentPlaceholder) isDeleted() bool {
	// A successful mapping the UpContentPlaceholder of a Delete event implies
	// no body transformation, therefore empty WebUrl attribute
	return p.WebURL == ""
}

type publicationEvent struct {
	ContentURI   string                `json:"contentUri"`
	Payload      *UpContentPlaceholder `json:"payload,omitempty"`
	LastModified string                `json:"lastModified"`
}

// MappingError is an error that can be returned by the content placeholder mapper
type MappingError struct {
	ContentUUID  string
	ErrorMessage string
}

func (e MappingError) Error() string {
	return e.ErrorMessage
}

// NewMappingError returs a new instance of a MappingError
func NewMappingError() *MappingError {
	return &MappingError{}
}

// WithMessage adds a message to a mapping error
func (e *MappingError) WithMessage(errorMsg string) *MappingError {
	e.ErrorMessage = errorMsg
	return e
}

// ForContent associate the mapping error to a specific piece of content
func (e *MappingError) ForContent(uuid string) *MappingError {
	e.ContentUUID = uuid
	return e
}
