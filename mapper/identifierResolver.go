package mapper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const (
	authorityPrefix = "http://api.ft.com/system/"
	uuidPattern     = "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$"
)

var uuidRegex = regexp.MustCompile(uuidPattern)

type IResolver interface {
	ResolveIdentifier(serviceId, refField, tid string) (string, error)
	CheckContentExists(uuid, tid string) error
}

type httpIResolver struct {
	brandMappings map[string]string
	client        DocStoreClient
}

func NewHttpIResolver(client DocStoreClient, brandMappings map[string]string) *httpIResolver {
	return &httpIResolver{client: client, brandMappings: brandMappings}
}

func (r *httpIResolver) ResolveIdentifier(serviceId, refField, tid string) (string, error) {
	mappingKey := strings.Split(serviceId, "?")[0]
	mappingKey = strings.Split(mappingKey, "#")[0]
	for key, value := range r.brandMappings {
		if strings.Contains(mappingKey, key) {
			authority := authorityPrefix + value
			identifierValue := strings.Split(serviceId, "://")[0] + "://" + key + "/?p=" + refField
			return r.resolveIdentifier(authority, identifierValue, tid)
		}
	}
	return "", fmt.Errorf("couldn't find authority in mapping table serviceId=%v refField=%v", serviceId, refField)
}

func (r *httpIResolver) resolveIdentifier(authority string, identifier string, tid string) (string, error) {
	status, location, err := r.client.ContentQuery(authority, identifier, tid)
	if err != nil {
		return "", err
	}
	if status != http.StatusMovedPermanently {
		return "", fmt.Errorf("unexpected response code while fetching canonical identifier for authority=%v identifier=%v status=%v", authority, identifier, status)
	}

	parts := strings.Split(location, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("resolved a canonical identifier which is an invalid FT URI for authority=%v identifier=%v location=%v", authority, identifier, location)
	}
	uuid := parts[len(parts)-1]
	if !uuidRegex.MatchString(uuid) {
		fmt.Println(parts)
		return "", fmt.Errorf("resolved a canonical identifier which contains an invalid uuid for authority=%v identifier=%v uuid=%v", authority, identifier, uuid)
	}

	return uuid, nil
}

func (r *httpIResolver) CheckContentExists(uuid, tid string) error {
	err := r.client.CheckContentExists(uuid, tid)
	if err != nil {
		return fmt.Errorf("failed to get content from document-store-api: %v", err.Error())
	}
	return nil
}
