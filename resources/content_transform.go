package resources

import (
	"encoding/json"
	"fmt"
	"net/http"

	tid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
)

// ContentTransformHandler is a HTTP handler to map methode content placeholders
type MapEndpointHandler struct {
	mapper mapper.Mapper
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
	methodePlaceholder, err := h.mapper.NewMethodeContentPlaceholderFromHTTPRequest(r)
	uuid := methodePlaceholder.UUID
	if err != nil {
		writeError(w, err, transactionID, uuid, r.RequestURI)
		return
	}
	upPlaceholder, err := h.mapper.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		writeError(w, err, transactionID, uuid, r.RequestURI)
		return
	}
	if contentHasBeenDeleted(upPlaceholder) {
		writeMessageForDeletedContent(w, transactionID, uuid, r.RequestURI)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(upPlaceholder)
	log.WithField("transaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", r.RequestURI).Info("Transformation successful")
}

func writeError(w http.ResponseWriter, err error, transactionID, uuid, requestURI string) {
	log.WithField("transaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", requestURI).WithError(err).Error(fmt.Sprintf("Returned HTTP status: %v", http.StatusUnprocessableEntity))
	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
}

func writeMessageForDeletedContent(w http.ResponseWriter, transactionID, uuid, requestURI string) {
	msg := fmt.Sprintf("Content has been deleted. transaction_id=\"%v\" uuid=\"%v\" request_uri=\"%v\"", transactionID, uuid, requestURI)
	log.Info(msg)
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(msg))
}

func contentHasBeenDeleted(placeholder mapper.UpContentPlaceholder) bool {
	return placeholder.WebURL == ""
}
