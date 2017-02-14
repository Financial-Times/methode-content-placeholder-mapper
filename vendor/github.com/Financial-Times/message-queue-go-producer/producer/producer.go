package producer

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const CONTENT_TYPE_HEADER = "application/vnd.kafka.binary.v1+json"
const CRLF = "\r\n"

//Interface for message producer - which writes to kafka through the proxy
type MessageProducer interface {
	SendMessage(string, Message) error
}

type DefaultMessageProducer struct {
	config MessageProducerConfig
	client *http.Client
}

//Configuration for message producer
type MessageProducerConfig struct {
	//proxy address
	Addr          string `json:"address"`
	Topic         string `json:"topic"`
	Authorization string `json:"authorization"`
}

//Message is the higher-level representation of messages from the queue: containing headers and message body
type Message struct {
	Headers map[string]string
	Body    string
}

//Message format required by Kafka-Proxy containing all the Messages
type MessageWithRecords struct {
	Records []MessageRecord `json:"records"`
}

//Message format required by Kafka-Proxy
type MessageRecord struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NewMessageProducer returns a producer instance
func NewMessageProducer(config MessageProducerConfig) MessageProducer {
	return NewMessageProducerWithHTTPClient(config, &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		}})
}

// NewMessageProducerWithHTTPClient returns a producer instance with specified http client instance
func NewMessageProducerWithHTTPClient(config MessageProducerConfig, httpClient *http.Client) MessageProducer {
	return &DefaultMessageProducer{config, httpClient}
}

func (p *DefaultMessageProducer) SendMessage(uuid string, message Message) (err error) {

	//concatenate message headers with message body to form a proper message string to save
	messageString := buildMessage(message)
	return p.SendRawMessage(uuid, messageString)
}

func (p *DefaultMessageProducer) SendRawMessage(uuid string, message string) (err error) {

	//encode in base64 and envelope the message
	envelopedMessage, err := envelopeMessage(uuid, message)
	if err != nil {
		return
	}

	//create request
	req, err := constructRequest(p.config.Addr, p.config.Topic, p.config.Authorization, envelopedMessage)

	//make request
	resp, err := p.client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("ERROR - executing request: %s", err.Error())
		return errors.New(errMsg)
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	//send - verify response status
	//log if error happens
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("ERROR - Unexpected response status %d. Expected: %d. %s", resp.StatusCode, http.StatusOK, resp.Request.URL.String())
		return errors.New(errMsg)
	}

	return nil
}

func constructRequest(addr string, topic string, authorizationKey string, message string) (*http.Request, error) {

	req, err := http.NewRequest("POST", addr+"/topics/"+topic, strings.NewReader(message))
	if err != nil {
		errMsg := fmt.Sprintf("ERROR - creating request: %s", err.Error())
		return req, errors.New(errMsg)
	}

	//set content-type header to json, and host header according to vulcand routing strategy
	req.Header.Add("Content-Type", CONTENT_TYPE_HEADER)

	if len(authorizationKey) > 0 {
		req.Header.Add("Authorization", authorizationKey)
	}

	return req, err
}

func buildMessage(message Message) string {

	builtMessage := "FTMSG/1.0" + CRLF

	var keys []string

	//order headers
	for header := range message.Headers {
		keys = append(keys, header)
	}
	sort.Strings(keys)

	//set headers
	for _, key := range keys {
		builtMessage = builtMessage + key + ": " + message.Headers[key] + CRLF
	}

	builtMessage = builtMessage + CRLF + message.Body

	return builtMessage

}

func envelopeMessage(key string, message string) (string, error) {

	var key64 string
	if key != "" {
		key64 = base64.StdEncoding.EncodeToString([]byte(key))
	}

	message64 := base64.StdEncoding.EncodeToString([]byte(message))

	record := MessageRecord{Key: key64, Value: message64}
	msgWithRecords := &MessageWithRecords{Records: []MessageRecord{record}}

	jsonRecords, err := json.Marshal(msgWithRecords)

	if err != nil {
		errMsg := fmt.Sprintf("ERROR - marshalling in json: %s", err.Error())
		return "", errors.New(errMsg)
	}

	return string(jsonRecords), err
}
