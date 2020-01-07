package mapper

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/Financial-Times/methode-content-placeholder-mapper/v2/model"
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
		return nil, fmt.Errorf("error unmarshalling methode messageBody: %v", err)
	}
	if p.Type != eomCompoundStory {
		return nil, model.NewInvalidMethodeCPH(fmt.Sprintf("Methode content has not type %s", eomCompoundStory))
	}

	attrs, err := buildAttributes(p.AttributesXML)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling not build xml attributes: %v", err.Error())
	}
	p.Attributes = attrs

	if p.Attributes.SourceCode != contentPlaceholderSourceCode {
		return nil, model.NewInvalidMethodeCPH("Methode content is not a content placeholder")
	}

	body, err := buildMethodeBody(p.Value)
	if err != nil {
		return nil, fmt.Errorf("error decoding or unmarshalling methode body: %v", err.Error())
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
	if err := xml.Unmarshal(methodeBodyXML, &body); err != nil {
		return model.MethodeBody{}, err
	}
	return body, nil
}
