package mapper

import (
	"errors"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"
)

const methodeSystemId = "methode-web-pub"
const methodeDateFormat = "20060102150405"
const upDateFormat = "2006-01-02T03:04:05.000Z0700"
const ftBrand = "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
const methodeAuthority = "http://api.ft.com/system/FTCOM-METHODE"
const ftApiContentUriPrefix = "http://api.ft.com/content/"

type Mapper interface {
	HandlePlaceholderMessages(msg consumer.Message)
	StartMappingMessages(c consumer.Consumer, p producer.MessageProducer)
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
	if msg.Headers["Origin-System-Id"] != methodeSystemId {
		log.WithField("transaction_id", tid).WithField("Origin-System-Id", msg.Headers["Origin-System-Id"]).Info("Ignoring message with different Origin-System-Id")
		return
	}
	placeholderMsg, placeholderUuid, err := m.mapMessage(msg)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Warn("Error in mapping message")
		return
	}
	err = m.messageProducer.SendMessage("", placeholderMsg)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Warn("Error sending transformed message to queue")
	}
	log.WithField("transaction_id", tid).WithField("uuid", placeholderUuid).Info("Content mapped and sent to the queue")
}

func (m *mapper) mapMessage(msg consumer.Message) (producer.Message, string, error) {
	methodePlaceholder, err := newMethodeContentPlaceholder(msg)
	if err != nil {
		return producer.Message{}, "", err
	}

	upPlaceholder, err := m.mapContentPlaceholder(methodePlaceholder)
	if err != nil {
		return producer.Message{}, "", err
	}

	upPlaceholderMsg, err := upPlaceholder.toProducerMessage()
	if err != nil {
		return producer.Message{}, "", err
	}

	return upPlaceholderMsg, upPlaceholder.UUID, nil
}

func (m *mapper) mapContentPlaceholder(mpc MethodeContentPlaceholder) (UpContentPlaceholder, error) {
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
		PublishReference:      mpc.transactionId,
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
	_, err := url.Parse(headline.Url)
	return err
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

func buildAlternativeTitles(promoTitle string) AlternativeTitles {
	if promoTitle == "" {
		return AlternativeTitles{}
	}
	return AlternativeTitles{PromotionalTitle: promoTitle}
}

func buildAlternativeImages(fileRef string) AlternativeImages {
	if fileRef == "" {
		return AlternativeImages{}
	}
	imageUuid := extractImageUuid(fileRef)
	return AlternativeImages{PromotionalImage: ftApiContentUriPrefix + imageUuid}
}

func extractImageUuid(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func buildAlternativeStandfirst(promoStandfirst string) AlternativeStandfirst {
	if promoStandfirst == "" {
		return AlternativeStandfirst{}
	}
	return AlternativeStandfirst{PromotionalStandfirst: promoStandfirst}
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

	log.Infof("Starting queue consumer: %#v", m.messageConsumer)
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
