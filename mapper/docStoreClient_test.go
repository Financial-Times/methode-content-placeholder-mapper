package mapper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/stretchr/testify/assert"
)

func successfulDocumentStoreServerMock(t *testing.T, filename string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentAsBytes, err := ioutil.ReadFile("test_resources/" + filename)
		if err != nil {
			assert.NoError(t, err, "Unable to open test resource file")
			return
		}
		w.Write(contentAsBytes)
	}))
}

func errorDocumentStoreServerMock(t *testing.T, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, http.StatusOK, statusCode, fmt.Sprintf("Status code should not be %d", http.StatusOK))
		w.WriteHeader(statusCode)
	}))
}

func TestGetContent_StatusOk(t *testing.T) {
	serverMock := successfulDocumentStoreServerMock(t, "document_store_content.json")
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)

	content, err := client.GetContent("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")

	assert.NoError(t, err, "Error wasn't expected during GetContent")
	assert.Equal(t, []model.Brand{{ID: "http://api.ft.com/things/164d0c3b-8a5a-4163-9519-96b57ed159bf"}, {ID: "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"}}, content.Brands)
	assert.Equal(t, "abcf2660-bbad-4a56-8eca-d0f8f0fac068", content.GetUppCoreContent().UUID)
	assert.Equal(t, "tid_Ml748dA0Wt_carousel_1509368354_gentx", content.GetUppCoreContent().PublishReference)
	assert.Equal(t, "2017-10-30T12:59:14.183Z", content.GetUppCoreContent().LastModified)
}

func TestGetContent_StatusInternalServerError(t *testing.T) {
	serverMock := errorDocumentStoreServerMock(t, http.StatusInternalServerError)
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)

	_, err := client.GetContent("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")

	assert.Error(t, err)
}

func TestGetContent_StatusNotFound(t *testing.T) {
	serverMock := errorDocumentStoreServerMock(t, http.StatusNotFound)
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)

	_, err := client.GetContent("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")

	assert.Error(t, err)
}

func TestGetContent_StatusServiceUnavailable(t *testing.T) {
	serverMock := errorDocumentStoreServerMock(t, http.StatusServiceUnavailable)
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)

	_, err := client.GetContent("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")

	assert.Error(t, err)
}

func TestGetContent_InvalidJsonInResponse(t *testing.T) {
	serverMock := successfulDocumentStoreServerMock(t, "invalid_document_store_content.json")
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)

	_, err := client.GetContent("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")

	assert.Error(t, err)
}

func TestCheckContentExists_StatusNotFound(t *testing.T) {
	serverMock := errorDocumentStoreServerMock(t, http.StatusNotFound)
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)
	err := client.CheckContentExists("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")
	assert.Error(t, err)
}

func TestCheckContentExists_StatusOK(t *testing.T) {
	serverMock := successfulDocumentStoreServerMock(t, "document_store_content.json")
	defer serverMock.Close()
	client := NewHttpDocStoreClient(http.DefaultClient, serverMock.URL)
	err := client.CheckContentExists("e1f02660-d41a-4a56-8eca-d0f8f0fac068", "tid_bh7VTFj9Il")
	assert.NoError(t, err)
}
