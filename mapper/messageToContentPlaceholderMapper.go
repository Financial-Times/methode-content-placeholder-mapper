package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

const contentPlaceholderSourceCode = "ContentPlaceholder"
const eomCompoundStory = "EOM::CompoundStory"

type MessageToContentPlaceholderMapper interface {
	Map(messageBody []byte) (*model.MethodeContentPlaceholder, error)
}

type DefaultMessageMapper struct {
}

func (m DefaultMessageMapper) Map(messageBody []byte) (*model.MethodeContentPlaceholder, error) {
	var p model.MethodeContentPlaceholder
	if err := json.Unmarshal(messageBody, &p); err != nil {
		return nil, fmt.Errorf("Error unmarshalling methode messageBody: %v", err)
	}
	if p.Type != eomCompoundStory {
		return nil, fmt.Errorf("Methode content has not type " + eomCompoundStory)
	}

	attrs, err := buildAttributes(p.AttributesXML)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling not build xml attributes.")
	}
	p.Attributes = attrs

	if p.Attributes.SourceCode != contentPlaceholderSourceCode {
		return nil, fmt.Errorf("Methode content is not a content placeholder")
	}

	body, err := buildMethodeBody(p.Value)
	if err != nil {
		return nil, fmt.Errorf("Error decoding or unmarshalling methode body.")
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
