package mapper

import (
	"errors"
	"net/url"

	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
)

type CPHValidator interface {
	Validate(mcp *model.MethodeContentPlaceholder) error
}

type defaultCPHValidator struct {
}

func NewDefaultCPHValidator() *defaultCPHValidator {
	return &defaultCPHValidator{}
}

func (dcv *defaultCPHValidator) Validate(mcp *model.MethodeContentPlaceholder) error {
	return dcv.validateHeadline(mcp.Body.LeadHeadline)
}

func (dcv *defaultCPHValidator) validateHeadline(headline model.LeadHeadline) error {
	if headline.Text == "" {
		return errors.New("Methode Content headline does not contain text")
	}
	if headline.URL == "" {
		return errors.New("Methode Content headline does not contain a link")
	}
	headlineURL, err := url.Parse(headline.URL)
	if err != nil {
		return errors.New("Methode Content headline does not contain a valid URL - " + err.Error())
	}
	if !headlineURL.IsAbs() {
		return errors.New("Methode Content headline does not contain an absolute URL")
	}
	return nil
}
