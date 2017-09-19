package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"strings"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type complementaryContentCPHMapper struct {
}

func (ccm *complementaryContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder) ([]model.UppContent, *utility.MappingError) {
	uuidToSet := mcp.UUID
	if mcp.IsInternalCPH() {
		uuidToSet = mcp.Attributes.LinkedArticleUUID
	}

	if mcp.Attributes.IsDeleted {
		return []model.UppContent{ccm.NewUppComplementaryContentDelete(mcp, uuidToSet)}, nil
	}

	return []model.UppContent{ccm.NewUppComplementaryContent(mcp, uuidToSet)}, nil
}

func (ccm *complementaryContentCPHMapper) NewUppComplementaryContent(mpc *model.MethodeContentPlaceholder, linkedArticleUUID string) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
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

func (ccm *complementaryContentCPHMapper) NewUppComplementaryContentDelete(mpc *model.MethodeContentPlaceholder, linkedArticleUUID string) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             linkedArticleUUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  true},
	}
}

func buildCCAlternativeTitles(promoTitle string) *model.AlternativeTitles {
	promoTitle = strings.TrimSpace(promoTitle)

	if promoTitle == "" {
		return nil
	}
	return &model.AlternativeTitles{PromotionalTitle: promoTitle}
}

func buildCCAlternativeImages(fileRef string) *model.AlternativeImages {
	if fileRef == "" {
		return nil
	}
	imageUUID := extractImageUUID(fileRef)
	return &model.AlternativeImages{PromotionalImage: imageUUID}
}

func extractImageUUID(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func buildCCAlternativeStandfirsts(promoStandfirst string) *model.AlternativeStandfirsts {
	promoStandfirst = strings.TrimSpace(promoStandfirst)
	if promoStandfirst == "" {
		return nil
	}
	return &model.AlternativeStandfirsts{PromotionalStandfirst: promoStandfirst}
}
