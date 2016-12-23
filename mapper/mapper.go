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
const eomCompandStory = "EOM::CompoundStory"

const upDateFormat = "2006-01-02T03:04:05.000Z0700"
const ftBrand = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
const methodeAuthority = "http://api.ft.com/system/FTCOM-METHODE"
const ftAPIContentURIPrefix = "http://api.ft.com/content/"
const mapperURIBase = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/content/"

type Mapper interface {
	HandlePlaceholderMessages(msg consumer.Message)
	StartMappingMessages(c consumer.Consumer, p producer.MessageProducer)
	NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (MethodeContentPlaceholder, error)
	MapContentPlaceholder(mpc MethodeContentPlaceholder) (UpContentPlaceholder, error)
}

type mapper struct {
	messageConsumer consumer.Consumer
	messageProducer producer.MessageProducer
}

func New() Mapper {
	return &mapper{}
}

func (m *mapper) HandlePlaceholderMessages(msg consumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if msg.Headers["Origin-System-Id"] != methodeSystemID {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}
	placeholderMsg, placeholderUUID, err := m.mapMessage(msg)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Warn("Error in mapping message")
		return
	}
	err = m.messageProducer.SendMessage("", placeholderMsg)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Warn("Error sending transformed message to queue")
	}
	log.WithField("transaction_id", tid).WithField("uuid", placeholderUUID).Info("Content mapped and sent to the queue")
}

func (m *mapper) mapMessage(msg consumer.Message) (producer.Message, string, error) {
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

func (m *mapper) MapContentPlaceholder(mpc MethodeContentPlaceholder) (UpContentPlaceholder, error) {
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
		return UpContentPlaceholder{}, err
	}

	publishDate, err := buildPublishedDate(mpc.attributes.LastPublicationDate)
	if err != nil {
		return UpContentPlaceholder{}, err
	}

	upPlaceholder := UpContentPlaceholder{
		UUID:                  mpc.UUID,
		Identifiers:           buildIdentifiers(mpc.UUID),
		Brands:                buildBrands(),
		WebUrl:                mpc.body.LeadHeadline.Url,
		AlternativeTitles:     buildAlternativeTitles(mpc.body.LeadHeadline.Text),
		AlternativeImages:     buildAlternativeImages(mpc.body.LeadImage.FileRef),
		AlternativeStandfirst: buildAlternativeStandfirst(mpc.body.LongStandfirst),
		PublishedDate:         publishDate,
		PublishReference:      mpc.transactionID,
		LastModified:          mpc.lastModified,
		CanBeSyndicated:       "verify",
	}
	return upPlaceholder, nil
}

func validateHeadline(headline LeadHeadline) error {
	if headline.Text == "" {
		return errors.New("Methode Content headline does not contain text")
	}
	if headline.Url == "" {
		return errors.New("Methode Content headline does not contain a link")
	}
	_, err := url.ParseRequestURI(headline.Url)
	//fmt.Println(url.Parse(headline.Url))
	if err != nil {
		return errors.New("Methode Content headline does not contain a valid URL - " + err.Error())
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
	//TODO check brands for IG and podcast
	brand := Brand{ID: ftBrand}
	return []Brand{brand}
}

func buildAlternativeTitles(promoTitle string) *AlternativeTitles {
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
	return &AlternativeImages{PromotionalImage: ftAPIContentURIPrefix + imageUUID}
}

func extractImageUUID(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func buildAlternativeStandfirst(promoStandfirst string) *AlternativeStandfirst {
	if promoStandfirst == "" {
		return nil
	}
	return &AlternativeStandfirst{PromotionalStandfirst: promoStandfirst}
}

func buildPublishedDate(lastPublicationDate string) (string, error) {
	date, err := time.Parse(methodeDateFormat, lastPublicationDate)
	if err != nil {
		return "", err
	}
	return date.Format(upDateFormat), nil
}

func (m *mapper) StartMappingMessages(c consumer.Consumer, p producer.MessageProducer) {
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

type Attributes struct {
	XMLName             xml.Name `xml:"ObjectMetadata"`
	SourceCode          string   `xml:"EditorialNotes>Sources>Source>SourceCode"`
	LastPublicationDate string   `xml:"OutputChannels>DIFTcom>DIFTcomLastPublication"`
	IsDeleted           bool     `xml:"OutputChannels>DIFTcom>DIFTcomMarkDeleted"`
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

func (m *mapper) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (MethodeContentPlaceholder, error) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	lastModified := time.Now().String()
	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return MethodeContentPlaceholder{}, err
	}
	return m.newMethodeContentPlaceholder(messageBody, transactionID, lastModified)
}

func (m *mapper) newMethodeContentPlaceholder(messageBody []byte, transactionID string, lastModified string) (MethodeContentPlaceholder, error) {
	var p MethodeContentPlaceholder
	if err := json.Unmarshal(messageBody, &p); err != nil {
		return MethodeContentPlaceholder{}, err
	}
	if p.Type != eomCompandStory {
		return MethodeContentPlaceholder{}, errors.New("Methode content has not type " + eomCompandStory)
	}

	p.transactionID = transactionID
	p.lastModified = lastModified

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

	if p.attributes.SourceCode != contentPlaceholderSourceCode {
		return MethodeContentPlaceholder{}, errors.New("Methode content is not a content placeholder")
	}
	return p, nil
}

func (m *mapper) newMethodeContentPlaceholderFromConsumerMessage(msg consumer.Message) (MethodeContentPlaceholder, error) {
	transactionID := msg.Headers["X-Request-Id"]
	lastModified := msg.Headers["Message-Timestamp"]

	return m.newMethodeContentPlaceholder([]byte(msg.Body), transactionID, lastModified)
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
	UUID                  string                 `json:"uuid"`
	Identifiers           []Identifier           `json:"identifiers"`
	Brands                []Brand                `json:"brands"`
	AlternativeTitles     *AlternativeTitles     `json:"alternativeTitles"`
	AlternativeImages     *AlternativeImages     `json:"alternativeImages"`
	AlternativeStandfirst *AlternativeStandfirst `json:"alternativeStandfirst"`
	PublishedDate         string                 `json:"publishedDate"`
	PublishReference      string                 `json:"publishReference"`
	LastModified          string                 `json:"lastModified"`
	WebUrl                string                 `json:"webUrl"`
	CanBeSyndicated       string                 `json:"canBeSyndicated"`
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

func (p UpContentPlaceholder) toPublicationEventMessage() (producer.Message, error) {

	publicationEvent := p.toPublicationEvent()

	jsonPublicationEvent, err := json.Marshal(publicationEvent)
	if err != nil {
		return producer.Message{}, err
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
	return p.WebUrl == ""
}

type publicationEvent struct {
	ContentURI   string                `json:"contentUri"`
	Payload      *UpContentPlaceholder `json:"payload,omitempty"`
	LastModified string                `json:"lastModified"`
}
