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
	ResolveIdentifier(serviceID, refField, tid string) (string, error)
	ContentExists(uuid, tid string) (bool, error)
}

type HTTPIResolver struct {
	brandMappings map[string]string
	client        DocStoreClient
}

func NewHttpIResolver(client DocStoreClient, brandMappings map[string]string) *HTTPIResolver {
	return &HTTPIResolver{client: client, brandMappings: brandMappings}
}

func (r *HTTPIResolver) ResolveIdentifier(serviceID, refField, tid string) (string, error) {
	mappingKey := strings.Split(serviceID, "?")[0]
	mappingKey = strings.Split(mappingKey, "#")[0]
	for key, value := range r.brandMappings {
		if strings.Contains(mappingKey, key) {
			authority := authorityPrefix + value
			identifierValue := strings.Split(serviceID, "://")[0] + "://" + key + "/?p=" + refField
			return r.resolveIdentifier(authority, identifierValue, tid)
		}
	}
	return "", fmt.Errorf("couldn't find authority in mapping table serviceId=%v refField=%v", serviceID, refField)
}

func (r *HTTPIResolver) resolveIdentifier(authority string, identifier string, tid string) (string, error) {
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

func (r *HTTPIResolver) ContentExists(uuid, tid string) (bool, error) {
	return r.client.ContentExists(uuid, tid)
}
