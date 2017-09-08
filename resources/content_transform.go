package resources

import (
	"encoding/json"
	"fmt"
	"net/http"

	tid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"time"
	"io/ioutil"
)

// ContentTransformHandler is a HTTP handler to map methode content placeholders
type MapEndpointHandler struct {
	mapper mapper.Mapper
}

type msg struct {
	Message string `json:"message"`
}

type mapResponse struct {
}

// NewContentTransformHandler returns a new instance of a MapEndpointHandler
func NewMapEndpointHandler(m mapper.Mapper) *MapEndpointHandler {
	return &MapEndpointHandler{m}
}

func (h *MapEndpointHandler) ServeMapEndpoint(w http.ResponseWriter, r *http.Request) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	log.WithField("transaction_id", transactionID).WithField("request_uri", r.RequestURI).Info("Received transformation request")
	h.mapContent(w, r, transactionID)
}

func (h *MapEndpointHandler) mapContent(w http.ResponseWriter, r *http.Request, transactionID string) {
	methodePlaceholder, err := h.NewMethodeContentPlaceholderFromHTTPRequest(r)

	if err != nil {
		writeError(w, err, transactionID, "", r.RequestURI)
		return
	}
	uuid := methodePlaceholder.UUID

	uppPlaceholder, uppComplementaryContent, err := h.mapper.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		writeError(w, err, transactionID, uuid, r.RequestURI)
		return
	}

	if uppPlaceholder.IsMarkedDeleted {
		writeMessageForDeletedContent(w, transactionID, uuid, r.RequestURI)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(h.createPublicationEventsMessages(uppPlaceholder, uppComplementaryContent))
	log.WithField("transaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", r.RequestURI).Info("Transformation successful")
}

func (h *MapEndpointHandler) NewMethodeContentPlaceholderFromHTTPRequest(r *http.Request) (*model.MethodeContentPlaceholder, *utility.MappingError) {
	transactionID := tid.GetTransactionIDFromRequest(r)
	lastModified := time.Now().Format(model.UPPDateFormat)
	messageBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error())
	}
	return model.NewMethodeContentPlaceholder(messageBody, transactionID, lastModified)
}

func (h *MapEndpointHandler) createPublicationEventsMessages(uppPlaceholder *model.UppContentPlaceholder, uppComplementaryContent *model.UppComplementaryContent) [2]model.PublicationEvent {
	return [2]model.PublicationEvent{
		{ContentURI: uppPlaceholder.ContentURI, Payload: uppPlaceholder, LastModified: uppPlaceholder.LastModified},
		{ContentURI: uppComplementaryContent.ContentURI, Payload: uppComplementaryContent, LastModified: uppComplementaryContent.LastModified}}
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
