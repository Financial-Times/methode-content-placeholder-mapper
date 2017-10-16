package resources

import (
	"encoding/json"
	"fmt"
	"net/http"

	tidUtils "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/message"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
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
		aggregateMapper:   aggregateMapper,
		cphMessageCreator: messageCreator,
		nativeMapper:      nativeMapper,
	}
}

func (h *MapEndpointHandler) ServeMapEndpoint(w http.ResponseWriter, r *http.Request) {
	tid := tidUtils.GetTransactionIDFromRequest(r)
	log.WithField("transaction_id", tid).WithField("request_uri", r.RequestURI).Info("Received transformation request")
	lmd := time.Now().Format(model.UPPDateFormat)

	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err, tid, "Could not read messageBody from request.", r.RequestURI)
		return
	}
	methodePlaceholder, err := h.nativeMapper.Map(messageBody)
	if err != nil {
		writeError(w, err, tid, "Could not map request body to intermediate model.", r.RequestURI)
	}

	if methodePlaceholder.Attributes.IsDeleted {
		writeMessageForDeletedContent(w, tid, methodePlaceholder.UUID, r.RequestURI)
		return
	}

	transformedContents, err := h.aggregateMapper.MapContentPlaceholder(methodePlaceholder, tid, lmd)
	if err != nil {
		writeError(w, err, tid, "Error mapping model from queue message.", r.RequestURI)
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
	log.WithField("transaction_id", tid).WithField("uuid", methodePlaceholder.UUID).WithField("request_uri", r.RequestURI).Info("Transformation successful")
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
