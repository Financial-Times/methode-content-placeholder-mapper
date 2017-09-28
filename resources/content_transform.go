package resources

import (
	"encoding/json"
	"fmt"
	"net/http"

	tid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"io/ioutil"
	"time"
)

type MapEndpointHandler struct {
	aggregateMapper   mapper.CPHAggregateMapper
	nativeMapper      mapper.MessageToContentPlaceholderMapper
	cphMessageCreator message.MessageCreator
}

type msg struct {
	Message string `json:"message"`
}

func NewMapEndpointHandler(aggregateMapper mapper.CPHAggregateMapper, messageCreator message.MessageCreator, nativeMapper mapper.MessageToContentPlaceholderMapper) *MapEndpointHandler {
	return &MapEndpointHandler{
		aggregateMapper: aggregateMapper,
		cphMessageCreator: messageCreator,
		nativeMapper: nativeMapper,
	}
}

func (h *MapEndpointHandler) ServeMapEndpoint(w http.ResponseWriter, r *http.Request) {
	tid := tid.GetTransactionIDFromRequest(r)
	log.WithField("transaction_id", tid).WithField("request_uri", r.RequestURI).Info("Received transformation request")
	h.mapContent(w, r, tid)
}

func (h *MapEndpointHandler) mapContent(w http.ResponseWriter, r *http.Request, tid string) {
	methodePlaceholder, err := h.NewMethodeContentPlaceholderFromHTTPRequest(r)

	if err != nil {
		writeError(w, err, tid, "could not get uuid from model", r.RequestURI)
		return
	}
	uuid := methodePlaceholder.UUID

	if methodePlaceholder.Attributes.IsDeleted {
		writeMessageForDeletedContent(w, tid, uuid, r.RequestURI)
		return
	}

	transformedContents, err := h.aggregateMapper.MapContentPlaceholder(methodePlaceholder, tid)
	if err != nil {
		log.WithField("transaction_id", tid).WithError(err).Error("Error mapping model from queue message")
		return
	}

	var pubEvents []model.PublicationEvent
	for _, transformedContent := range transformedContents {
		pubEvent := h.cphMessageCreator.ToPublicationEvent(transformedContent.GetUppCoreContent(), transformedContent)
		pubEvents = append(pubEvents, *pubEvent)
	}

	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(pubEvents)
	log.WithField("transaction_id", tid).WithField("uuid", uuid).WithField("request_uri", r.RequestURI).Info("Transformation successful")
}

func (h *MapEndpointHandler) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (*model.MethodeContentPlaceholder, *utility.MappingError) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	lastModified := time.Now().Format(model.UPPDateFormat)
	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error())
	}
	return h.nativeMapper.Map(messageBody, transactionID, lastModified)
}

func writeError(w http.ResponseWriter, err error, transactionID, uuid, requestURI string) {
	log.WithField("transaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", requestURI).WithError(err).Error(fmt.Sprintf("Returned HTTP status: %v", http.StatusUnprocessableEntity))
	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
}

func writeMessageForDeletedContent(w http.ResponseWriter, transactionID, uuid, requestURI string) {
	log.WithField("transaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", requestURI).Info("Content has been deleted.")
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-Request-ID", transactionID)

	w.WriteHeader(http.StatusNotFound)

	data, _ := json.Marshal(&msg{Message: "Delete event"})
	w.Write(data)
}
