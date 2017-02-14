package resources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/stretchr/testify/assert"
)

const mockedTopics = `["methode-articles","up-placholders"]`

var consumerConfigMock = consumer.QueueConfig{
	Group:            "mcpm-group",
	Topic:            "methode-articles",
	AuthorizationKey: "my-first-auth-key",
}

var producerConfigMock = producer.MessageProducerConfig{
	Topic:         "up-placholders",
	Authorization: "my-first-auth-key",
}

func setupMockKafka(t *testing.T, status int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if status != 200 {
			w.WriteHeader(status)
		} else {
			w.Write([]byte(response))
		}

		assert.Equal(t, "my-first-auth-key", req.Header.Get("Authorization"))
	}))
}

func TestHealthchecks(t *testing.T) {
	kafka := setupMockKafka(t, 200, mockedTopics)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	t.Log(len(result.Checks))
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.BusinessImpact != "")
	assert.Equal(t, "ConsumerQueueProxyReachable", consumerCheck.Name)
	assert.True(t, consumerCheck.Ok)
	assert.True(t, result.Ok)
	assert.True(t, consumerCheck.PanicGuide != "")
	assert.Equal(t, uint8(1), consumerCheck.Severity)

	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.BusinessImpact != "")
	assert.Equal(t, "ProducerQueueProxyReachable", producerCheck.Name)
	assert.True(t, producerCheck.Ok)
	assert.True(t, result.Ok)
	assert.True(t, producerCheck.PanicGuide != "")
	assert.Equal(t, uint8(1), producerCheck.Severity)
}

func TestTopicMissing(t *testing.T) {
	kafka := setupMockKafka(t, 200, `[]`)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
}

func TestTopicsUnparseable(t *testing.T) {
	kafka := setupMockKafka(t, 200, ``)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
}

func TestFailingKafka(t *testing.T) {
	kafka := setupMockKafka(t, 500, ``)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)
	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
}

func TestNoKafkaAtAll(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{"a-fake-url"}
	producerConfigMock.Addr = "a-fake-url"
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
}

func TestNoKafkaConsumer(t *testing.T) {
	kafka := setupMockKafka(t, 200, mockedTopics)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{"a-fake-url"}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
}

func TestNoKafkaProducer(t *testing.T) {
	kafka := setupMockKafka(t, 200, mockedTopics)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = "a-fake-url"
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.False(t, producerCheck.Ok)
}

func TestMultipleKafkaConsumersFail(t *testing.T) {
	kafka := setupMockKafka(t, 200, mockedTopics)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL, "a-fake-url"}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.False(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.False(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
}

func TestMultipleKafkaConsumersOK(t *testing.T) {
	kafka1 := setupMockKafka(t, 200, mockedTopics)
	defer kafka1.Close()
	kafka2 := setupMockKafka(t, 200, mockedTopics)
	defer kafka2.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/__health", nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka1.URL, kafka2.URL}
	producerConfigMock.Addr = kafka1.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.ConsumerQueueCheck(), hc.ProducerQueueCheck())(w, req)

	assert.Equal(t, 200, w.Code)

	decoder := json.NewDecoder(w.Body)

	var result fthealth.HealthResult
	decoder.Decode(&result)

	assert.True(t, result.Ok)
	consumerCheck := result.Checks[0]
	assert.True(t, consumerCheck.Ok)
	producerCheck := result.Checks[1]
	assert.True(t, producerCheck.Ok)
}

func TestGTG(t *testing.T) {
	kafka := setupMockKafka(t, 200, mockedTopics)
	defer kafka.Close()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", httphandlers.GTGPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	consumerConfigMock.Addrs = []string{kafka.URL}
	producerConfigMock.Addr = kafka.URL
	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	hc.GTG(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGTGFailing(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", httphandlers.GTGPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	hc := NewMapperHealthcheck(consumerConfigMock, producerConfigMock)
	hc.GTG(w, req)

	assert.Equal(t, 503, w.Code)
}
