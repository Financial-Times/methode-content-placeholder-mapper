package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
)

const contentPlaceholderSourceCode = "ContentPlaceholder"
const eomCompoundStory = "EOM::CompoundStory"

type MessageToContentPlaceholderMapper interface {
	Map(messageBody []byte, transactionID string, lastModified string) (*model.MethodeContentPlaceholder, *utility.MappingError)
}

type DefaultMessageMapper struct {
}

func (DefaultMessageMapper) Map(messageBody []byte, transactionID string, lastModified string) (*model.MethodeContentPlaceholder, *utility.MappingError) {
	var p model.MethodeContentPlaceholder
	if err := json.Unmarshal(messageBody, &p); err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error())
	}
	if p.Type != eomCompoundStory {
		return nil, utility.NewMappingError().WithMessage("Methode content has not type " + eomCompoundStory).ForContent(p.UUID)
	}

	p.TransactionID = transactionID
	p.LastModified = lastModified

	attrs, err := buildAttributes(p.AttributesXML)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.Attributes = attrs

	if p.Attributes.SourceCode != contentPlaceholderSourceCode {
		return nil, utility.NewMappingError().WithMessage("Methode content is not a content placeholder").ForContent(p.UUID)
	}

	body, err := buildMethodeBody(p.Value)
	if err != nil {
		return nil, utility.NewMappingError().WithMessage(err.Error()).ForContent(p.UUID)
	}
	p.Body = body

	return &p, nil
}

func buildAttributes(attributesXML string) (model.Attributes, error) {
	var attrs model.Attributes
	if err := xml.Unmarshal([]byte(attributesXML), &attrs); err != nil {
		return model.Attributes{}, err
	}
	return attrs, nil
}

func buildMethodeBody(methodeBodyXMLBase64 string) (model.MethodeBody, error) {
	methodeBodyXML, err := base64.StdEncoding.DecodeString(methodeBodyXMLBase64)
	if err != nil {
		return model.MethodeBody{}, err
	}
	var body model.MethodeBody
	if err := xml.Unmarshal([]byte(methodeBodyXML), &body); err != nil {
		return model.MethodeBody{}, err
	}
	return body, nil
}
