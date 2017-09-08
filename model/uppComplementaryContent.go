package model

import (
	"strings"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type UppComplementaryContent struct {
	UppCoreContent
	PromotionalTitle      string `json:"promotionalTitle,omitempty"`
	PromotionalImage      string `json:"promotionalImage"`
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}

func NewUppComplementaryContent(mpc *MethodeContentPlaceholder, linkedArticleUUID string) *UppComplementaryContent {
	return &UppComplementaryContent{
		UppCoreContent: UppCoreContent{
			UUID:             linkedArticleUUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted},
		PromotionalTitle:      buildPromotionalTitle(mpc.Body.LeadHeadline.Text),
		PromotionalImage:      buildPromotionalImage(mpc.Body.LeadImage.FileRef),
		PromotionalStandfirst: buildPromotionalStandfirst(mpc.Body.LongStandfirst),
	}
}

func NewUppComplementaryContentDelete(mpc *MethodeContentPlaceholder) *UppComplementaryContent {
	return &UppComplementaryContent{
		UppCoreContent: UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  true},
	}
}

func buildPromotionalTitle(promoTitle string) string {
	return strings.TrimSpace(promoTitle)
}

func buildPromotionalImage(fileRef string) string {
	if fileRef == "" {
		return fileRef
	}

	return strings.Split(fileRef, "uuid=")[1]
}

func buildPromotionalStandfirst(promoStandfirst string) string {
	return strings.TrimSpace(promoStandfirst)
}
