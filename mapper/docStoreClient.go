package mapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Sirupsen/logrus"
)

const documentStoreApiHost = "document-store-api"

type DocStoreClient interface {
	ContentQuery(authority, identifier, tid string) (status int, location string, err error)
	GetContent(uuid, tid string) (*model.DocStoreUppContent, error)
	ConnectivityCheck() (string, error)
}

type httpDocStoreClient struct {
	docStoreAddress string
	client          *http.Client
}

func NewHttpDocStoreClient(client *http.Client, docStoreAddress string) *httpDocStoreClient {
	return &httpDocStoreClient{docStoreAddress: docStoreAddress, client: client}
}

func (c *httpDocStoreClient) GetContent(uuid, tid string) (*model.DocStoreUppContent, error) {
	docStoreUrl, err := url.Parse(c.docStoreAddress + "/content/" + uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rawurl into URL structure for docStoreAddress=%v uuid=%v: %v", c.docStoreAddress, uuid, err.Error())
	}
	req, err := http.NewRequest(http.MethodGet, docStoreUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to fetch content for uuid=%v: %v", uuid, err.Error())
	}
	req.Host = documentStoreApiHost
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsucessful request for content for uuid=%v: %v", uuid, err.Error())
	}
	defer niceClose(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code=%v for uuid=%v", resp.StatusCode, uuid)
	}
	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for uuid=%v: %v", uuid, err.Error())
	}
	var content model.DocStoreUppContent
	err = json.Unmarshal(bodyAsBytes, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body for uuid=%v: %v", uuid, err.Error())
	}

	return &content, nil
}

func (c *httpDocStoreClient) ContentQuery(authority, identifier, tid string) (status int, location string, err error) {
	docStoreUrl, err := url.Parse(c.docStoreAddress + "/content-query")
	if err != nil {
		return -1, "", fmt.Errorf("Invalid address docStoreAddress=%v", c.docStoreAddress)
	}
	query := url.Values{}
	query.Add("identifierValue", identifier)
	query.Add("identifierAuthority", authority)
	docStoreUrl.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodGet, docStoreUrl.String(), nil)
	if err != nil {
		return -1, "", fmt.Errorf("Couldn't create request to fetch canonical identifier for authority=%v identifier=%v", authority, identifier)
	}
	// TODO: Remove when host based routing doesn't exist any more.
	req.Host = documentStoreApiHost
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.client.Do(req)
	if err != nil {
		return -1, "", fmt.Errorf("Unsucessful request for fetching canonical identifier for authority=%v identifier=%v", authority, identifier)
	}
	niceClose(resp)

	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func (c *httpDocStoreClient) ConnectivityCheck() (string, error) {
	docStoreGtgUrl, err := url.Parse(c.docStoreAddress + "/__gtg")
	if err != nil {
		return "Error connecting to document-store-api", fmt.Errorf("Invalid address docStoreAddress=%v", c.docStoreAddress)
	}
	req, err := http.NewRequest(http.MethodGet, docStoreGtgUrl.String(), nil)
	if err != nil {
		return "Error connecting to document-store-api", fmt.Errorf("Couldn't create request to GTG")
	}
	req.Host = "document-store-api"
	resp, err := c.client.Do(req)
	if err != nil {
		return "Error connecting to document-store-api", fmt.Errorf("Unsucessful request for GTG")
	}
	niceClose(resp)
	if resp.StatusCode != http.StatusOK {
		return "Error connecting to document-store-api", fmt.Errorf("status=%v", resp.StatusCode)
	}
	return "OK", nil
}

func niceClose(resp *http.Response) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logrus.Warnf("Couldn't close response body %v", err)
		}
	}()
}
