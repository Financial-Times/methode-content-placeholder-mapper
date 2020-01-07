package mapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/Financial-Times/methode-content-placeholder-mapper/v2/model"
	"github.com/sirupsen/logrus"
)

type DocStoreClient interface {
	ContentQuery(authority, identifier, tid string) (status int, location string, err error)
	GetContent(uuid, tid string) (*model.DocStoreUppContent, error)
	ContentExists(uuid, tid string) (bool, error)
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
	docStoreURL, err := url.Parse(c.docStoreAddress + "/content/" + uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rawurl into URL structure for docStoreAddress=%v uuid=%v: %v", c.docStoreAddress, uuid, err.Error())
	}
	req, err := http.NewRequest(http.MethodGet, docStoreURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to fetch content for uuid=%v: %v", uuid, err.Error())
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unsuccessful request for content for uuid=%v: %v", uuid, err.Error())
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

func (c *httpDocStoreClient) ContentExists(uuid, tid string) (bool, error) {
	docStoreUrl, err := url.Parse(c.docStoreAddress + "/content/" + uuid)
	if err != nil {
		return false, fmt.Errorf("failed to parse rawurl into URL structure for docStoreAddress=%v uuid=%v: %v", c.docStoreAddress, uuid, err.Error())
	}
	req, err := http.NewRequest(http.MethodGet, docStoreUrl.String(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request to fetch content for uuid=%v: %v", uuid, err.Error())
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("unsuccessful request for content for uuid=%v: %v", uuid, err.Error())
	}
	defer niceClose(resp)
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	return true, nil
}

func (c *httpDocStoreClient) ContentQuery(authority, identifier, tid string) (status int, location string, err error) {
	docStoreURL, err := url.Parse(c.docStoreAddress + "/content-query")
	if err != nil {
		return -1, "", fmt.Errorf("invalid address docStoreAddress=%v: %v", c.docStoreAddress, err.Error())
	}
	query := url.Values{}
	query.Add("identifierValue", identifier)
	query.Add("identifierAuthority", authority)
	docStoreURL.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodGet, docStoreURL.String(), nil)
	if err != nil {
		return -1, "", fmt.Errorf("couldn't create request to fetch canonical identifier for authority=%v identifier=%v: %v", authority, identifier, err.Error())
	}
	req.Header.Add(transactionidutils.TransactionIDHeader, tid)
	resp, err := c.client.Do(req)
	if err != nil {
		return -1, "", fmt.Errorf("unsuccessful request for fetching canonical identifier for authority=%v identifier=%v: %v", authority, identifier, err.Error())
	}
	niceClose(resp)

	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func (c *httpDocStoreClient) ConnectivityCheck() (string, error) {
	errMsg := "Error connecting to document-store-api"
	docStoreGtgUrl, err := url.Parse(c.docStoreAddress + "/__gtg")
	if err != nil {
		return errMsg, fmt.Errorf("invalid address docStoreAddress=%v: %v", c.docStoreAddress, err.Error())
	}
	req, err := http.NewRequest(http.MethodGet, docStoreGtgUrl.String(), nil)
	if err != nil {
		return errMsg, fmt.Errorf("couldn't create request to GTG: %v", err.Error())
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return errMsg, fmt.Errorf("unsuccessful request for GTG: %v", err.Error())
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
