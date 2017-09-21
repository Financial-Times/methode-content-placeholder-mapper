package mapper

import (
	"net/http"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/url"
	"net/http/httputil"
)

type docStoreClient interface {
	contentQuery(authority string, identifier string) (status int, location string, err error)
}

type httpDocStoreClient struct {
	docStoreAddress string
	client          *http.Client
}

func NewHttpDocStoreClient(client *http.Client, docStoreAddress string) *httpDocStoreClient {
	return &httpDocStoreClient{docStoreAddress: docStoreAddress, client: client}
}

func (c *httpDocStoreClient) contentQuery(authority string, identifier string) (status int, location string, err error) {
	docStoreUrl, err := url.Parse(c.docStoreAddress)
	if err != nil {
		return -1, "", fmt.Errorf("Invalid address docStoreAddress=%v", c.docStoreAddress)
	}
	docStoreUrl.Path += "content-query"
	parameters := url.Values{}
	parameters.Add("identifierValue", identifier)
	parameters.Add("identifierAuthority", authority)
	docStoreUrl.RawQuery = parameters.Encode()
	logrus.Infof("docStoreUrl.String()=%v", docStoreUrl.String())

	req, err := http.NewRequest("GET", docStoreUrl.String(), nil)
	logrus.Infof("req.URL.String()=%v", req.URL.String())
	if err != nil {
		return -1, "", fmt.Errorf("Couldn't create request to fetch canonical identifier for authority=%v identifier=%v", authority, identifier)
	}
	// TODO: Remove when host based routing doesn't exist any more.
	req.Host = "document-store-api"
	resp, err := c.client.Do(req)
	if err != nil {
		return -1, "", fmt.Errorf("Unsucessful request for fetching canonical identifier for authority=%v identifier=%v", authority, identifier)
	}
	niceClose(resp)
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logrus.Infof("dumping doesn't work")
	} else {
		logrus.Infof(string(dump))
	}

	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func niceClose(resp *http.Response) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logrus.Warnf("Couldn't close response body %v", err)
		}
	}()
}
