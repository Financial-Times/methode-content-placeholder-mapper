package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/utility"
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"strings"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type ComplementaryContentCPHMapper struct {
}

func (ccm *ComplementaryContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder, uuid string, tid string) ([]model.UppContent, *utility.MappingError) {
	var cc *model.UppComplementaryContent
	if mcp.Attributes.IsDeleted {
		cc = ccm.mapToUppComplementaryContentDelete(mcp)
	} else{
		cc = ccm.mapToUppComplementaryContent(mcp)
	}
	if uuid != "" {
		cc.UUID = uuid
	}
	return []model.UppContent{cc}, nil
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContent(mpc *model.MethodeContentPlaceholder) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted,
		},
		AlternativeTitles:      buildCCAlternativeTitles(mpc.Body.LeadHeadline.Text),
		AlternativeImages:      buildCCAlternativeImages(mpc.Body.LeadImage.FileRef),
		AlternativeStandfirsts: buildCCAlternativeStandfirsts(mpc.Body.LongStandfirst),
	}
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContentDelete(mpc *model.MethodeContentPlaceholder) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			PublishReference: mpc.TransactionID,
			LastModified:     mpc.LastModified,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  true,
		},
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
