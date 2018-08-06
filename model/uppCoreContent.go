package model

type UppContent interface {
	GetUUID() string
	GetUppCoreContent() *UppCoreContent
}

type UppCoreContent struct {
	UUID              string 			`json:"uuid"`
	PublishReference  string 			`json:"publishReference"`
	LastModified      string 			`json:"lastModified"`
	ContentURI        string 			`json:"-"`
	IsMarkedDeleted   bool   			`json:"-"`
	ContentTypeHeader map[string]string `json:"contentTypeHeader"`
}

func (ucp *UppCoreContent) GetUUID() string {
	return ucp.UUID
}

func (ucp *UppCoreContent) GetUppCoreContent() *UppCoreContent {
	return ucp
}
