package mapper

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Financial-Times/transactionid-utils-go"
	"github.com/Sirupsen/logrus"
)

type DocStoreClient interface {
	ContentQuery(authority string, identifier string, tid string) (status int, location string, err error)
	ConnectivityCheck() (string, error)
}

type httpDocStoreClient struct {
	docStoreAddress string
	client          *http.Client
}

func NewHttpDocStoreClient(client *http.Client, docStoreAddress string) *httpDocStoreClient {
	return &httpDocStoreClient{docStoreAddress: docStoreAddress, client: client}
}

func (c *httpDocStoreClient) ContentQuery(authority string, identifier string, tid string) (status int, location string, err error) {
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
	req.Host = "document-store-api"
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
	if resp.StatusCode != 200 {
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
