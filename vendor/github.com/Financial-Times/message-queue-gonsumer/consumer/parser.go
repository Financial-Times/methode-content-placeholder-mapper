package consumer

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

//raw message
type message struct {
	Value     string `json:"value"` //base64 encoded
	Partition int    `json:"partition"`
	Offset    int    `json:"offset"`
}

func parseResponse(data []byte) ([]Message, error) {
	var resp []message
	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Printf("ERROR - parsing json message %q failed with error %v", data, err.Error())
		return nil, err
	}
	var msgs []Message
	for _, m := range resp {
		//log.Printf("DEBUG - parsing msg of partition %d and offset %d", m.Partition, m.Offset)
		if msg, err := parseMessage(m.Value); err == nil {
			msgs = append(msgs, msg)
		} else {
			log.Printf("ERROR - parsing message %v", err.Error())
		}
	}
	return msgs, nil
}

// FT async msg format:
//
// message-version CRLF
// *(message-header CRLF)
// CRLF
// message-body
func parseMessage(raw string) (m Message, err error) {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		log.Printf("ERROR - failure in decoding base64 value: %s", err.Error())
		return
	}
	doubleNewLineStartIndex := getHeaderSectionEndingIndex(string(decoded[:]))
	if m.Headers, err = parseHeaders(string(decoded[:doubleNewLineStartIndex])); err != nil {
		return
	}
	m.Body = strings.TrimSpace(string(decoded[doubleNewLineStartIndex:]))
	return
}

func getHeaderSectionEndingIndex(msg string) int {
	//FT msg format uses CRLF for line endings
	i := strings.Index(msg, "\r\n\r\n")
	if i != -1 {
		return i
	}
	//fallback to UNIX line endings
	i = strings.Index(msg, "\n\n")
	if i != -1 {
		return i
	}
	log.Printf("WARN  - message with no message body: [%s]", msg)
	return len(msg)
}

var re = regexp.MustCompile("[\\w-]*:[\\w\\-:/. ]*")

var kre = regexp.MustCompile("[\\w-]*:")
var vre = regexp.MustCompile(":[\\w-:/. ]*")

func parseHeaders(msg string) (map[string]string, error) {
	headerLines := re.FindAllString(msg, -1)

	headers := make(map[string]string)
	for _, line := range headerLines {
		key, value := parseHeader(line)
		headers[key] = value
	}
	return headers, nil
}

func parseHeader(header string) (string, string) {
	key := kre.FindString(header)
	value := vre.FindString(header)
	return key[:len(key)-1], strings.TrimSpace(value[1:])
}
