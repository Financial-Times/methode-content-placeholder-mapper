package mapper

import (
	"net/http"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/url"
	"io/ioutil"
)

type docStoreClient interface {
	contentQuery(authority string, identifier string) (status int, location string, err error)
}

type httpDocStoreClient struct {
	docStoreAddress string
	client          *http.Client
}

func NewHttpDocStoreClient(client *http.Client, docStoreHostAndPort string) *httpDocStoreClient {
	return &httpDocStoreClient{docStoreAddress: docStoreHostAndPort, client: client}
}

func (c *httpDocStoreClient) contentQuery(authority string, identifier string) (status int, location string, err error) {
	docStoreUrl, err := url.Parse(c.docStoreAddress)
	if err != nil {
		return -1, "", fmt.Errorf("Invalid address docStoreAddress=%v", c.docStoreAddress)
	}
	docStoreUrl.Path += "content-query"
	parameters := url.Values{}
	parameters.Add("identifierAuthority", authority)
	parameters.Add("identifierValue", identifier)
	docStoreUrl.RawQuery = parameters.Encode()
	fmt.Println(docStoreUrl.String())
	req, err := http.NewRequest("GET", docStoreUrl.String(), nil)
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

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

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
