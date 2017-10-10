package mapper

import (
	"github.com/Financial-Times/methode-content-placeholder-mapper/model"
	"strings"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type ComplementaryContentCPHMapper struct {
}

func (ccm *ComplementaryContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder, uuid, tid, lmd string) ([]model.UppContent, error) {
	var cc *model.UppComplementaryContent
	if mcp.Attributes.IsDeleted {
		cc = ccm.mapToUppComplementaryContentDelete(mcp, tid, lmd)
	} else {
		cc = ccm.mapToUppComplementaryContent(mcp, tid, lmd)
	}
	if uuid != "" {
		cc.UUID = uuid
	}
	return []model.UppContent{cc}, nil
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContent(mpc *model.MethodeContentPlaceholder, tid, lmd string) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  mpc.Attributes.IsDeleted,
			PublishReference: tid,
			LastModified:     lmd,
		},
		Brands:                 buildBrands(),
		AlternativeTitles:      buildCCAlternativeTitles(mpc.Body.LeadHeadline.Text),
		AlternativeImages:      buildCCAlternativeImages(mpc.Body.LeadImage.FileRef),
		AlternativeStandfirsts: buildCCAlternativeStandfirsts(mpc.Body.LongStandfirst),
	}
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContentDelete(mpc *model.MethodeContentPlaceholder, tid, lmd string) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  true,
			PublishReference: tid,
			LastModified:     lmd,
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
