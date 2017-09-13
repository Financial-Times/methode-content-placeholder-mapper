package model

import (
	"strings"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type UppComplementaryContent struct {
	UppCoreContent
	AlternativeTitles      *AlternativeTitles      `json:"alternativeTitles"`
	AlternativeImages      *AlternativeImages      `json:"alternativeImages"`
	AlternativeStandfirsts *AlternativeStandfirsts `json:"alternativeStandfirsts"`
}

func NewUppComplementaryContent(mpc *MethodeContentPlaceholder, linkedArticleUUID string) *UppComplementaryContent {
	return &UppComplementaryContent{
		UppCoreContent: UppCoreContent{
			UUID:             linkedArticleUUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted},
		AlternativeTitles:      buildCCAlternativeTitles(mpc.Body.LeadHeadline.Text),
		AlternativeImages:      buildCCAlternativeImages(mpc.Body.LeadImage.FileRef),
		AlternativeStandfirsts: buildCCAlternativeStandfirsts(mpc.Body.LongStandfirst),
	}
}

func NewUppComplementaryContentDelete(mpc *MethodeContentPlaceholder, linkedArticleUUID string) *UppComplementaryContent {
	return &UppComplementaryContent{
		UppCoreContent: UppCoreContent{
			UUID:             linkedArticleUUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  true},
	}
}

type AlternativeImages struct {
	PromotionalImage string `json:"promotionalImage"`
}

type AlternativeStandfirsts struct {
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}

func buildCCAlternativeTitles(promoTitle string) *AlternativeTitles {
	promoTitle = strings.TrimSpace(promoTitle)

	if promoTitle == "" {
		return nil
	}
	return &AlternativeTitles{PromotionalTitle: promoTitle}
}

func buildCCAlternativeImages(fileRef string) *AlternativeImages {
	if fileRef == "" {
		return nil
	}
	imageUUID := extractImageUUID(fileRef)
	return &AlternativeImages{PromotionalImage: imageUUID}
}

func extractImageUUID(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func buildCCAlternativeStandfirsts(promoStandfirst string) *AlternativeStandfirsts {
	promoStandfirst = strings.TrimSpace(promoStandfirst)
	if promoStandfirst == "" {
		return nil
	}
	return &AlternativeStandfirsts{PromotionalStandfirst: promoStandfirst}
}
