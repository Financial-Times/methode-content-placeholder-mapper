package model

type PublicationEvent struct {
	ContentURI   string      			`json:"contentUri"`
	Payload      interface{} 			`json:"payload,omitempty"`
	LastModified string      			`json:"lastModified"`
	ContentTypeHeader map[string]string `json:"contentTypeHeader"`
}
