package resources

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Financial-Times/methode-content-placeholder-mapper/mapper"
	tid "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type ContentTransformHandler struct {
	mapper mapper.Mapper
}

func NewContentTransformHandler(m mapper.Mapper) *ContentTransformHandler {
	return &ContentTransformHandler{m}
}

func (h *ContentTransformHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	transactionID := tid.GetTransactionIDFromRequest(r)
	log.WithField("trasaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", r.RequestURI).Info("Received transformation request")
	methodePlaceholder, err := h.mapper.NewMethodeContentPlaceholderFromHTTPRequest(r)
	if err != nil {
		writeError(w, err, transactionID, uuid, r.RequestURI)
		return
	}
	upPlaceholder, err := h.mapper.MapContentPlaceholder(methodePlaceholder)
	if err != nil {
		writeError(w, err, transactionID, uuid, r.RequestURI)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(upPlaceholder)
	log.WithField("trasaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", r.RequestURI).Info("Transformation successful")
}

func writeError(w http.ResponseWriter, err error, transactionID string, uuid string, requestURI string) {
	log.WithField("trasaction_id", transactionID).WithField("uuid", uuid).WithField("request_uri", requestURI).WithError(err).Warn(fmt.Sprintf("Returned HTTP status: %v", http.StatusUnprocessableEntity))
	http.Error(w, err.Error(), http.StatusUnprocessableEntity)
}
