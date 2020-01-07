package mapper

import (
	"fmt"
	"strings"

	"github.com/Financial-TimesFinancial-Times/methode-content-placeholder-mapper/v2/model"
)

const complementaryContentURI = "http://methode-content-placeholder-mapper-iw-uk-p.svc.ft.com/complementarycontent/"

type ComplementaryContentCPHMapper struct {
	apiHostFormat string
	client        DocStoreClient
}

func NewComplementaryContentCPHMapper(apiHost string, client DocStoreClient) *ComplementaryContentCPHMapper {
	return &ComplementaryContentCPHMapper{
		apiHostFormat: "http://" + apiHost + "/content/%s",
		client:        client,
	}
}

func (ccm *ComplementaryContentCPHMapper) MapContentPlaceholder(mcp *model.MethodeContentPlaceholder, uuid, tid, lmd string) ([]model.UppContent, error) {
	var cc *model.UppComplementaryContent

	isInternalCPH := false
	markIfDelete := true
	if uuid != "" {
		isInternalCPH = true
		markIfDelete = false

	}

	if mcp.Attributes.IsDeleted {
		cc = ccm.mapToUppComplementaryContentDelete(mcp, tid, lmd, markIfDelete)
	} else {
		cc = ccm.mapToUppComplementaryContentUpdate(mcp, tid, lmd)
	}

	if isInternalCPH {
		cc.UUID = uuid
		if err := ccm.setBrands(uuid, tid, cc); err != nil {
			return nil, fmt.Errorf("failed to retrieve brands for complementary content: %v", err.Error())
		}
	}

	return []model.UppContent{cc}, nil
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContentUpdate(mpc *model.MethodeContentPlaceholder, tid, lmd string) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  false,
			PublishReference: tid,
			LastModified:     lmd,
		},
		Type:                   contentType,
		Brands:                 model.BuildBrands(),
		AlternativeTitles:      ccm.buildCCAlternativeTitles(mpc.Body.LeadHeadline.Text),
		AlternativeImages:      ccm.buildCCAlternativeImages(mpc.Body.LeadImage.FileRef),
		AlternativeStandfirsts: ccm.buildCCAlternativeStandfirsts(mpc.Body.LongStandfirst),
	}
}

func (ccm *ComplementaryContentCPHMapper) mapToUppComplementaryContentDelete(mpc *model.MethodeContentPlaceholder, tid, lmd string, markDelete bool) *model.UppComplementaryContent {
	return &model.UppComplementaryContent{
		UppCoreContent: model.UppCoreContent{
			UUID:             mpc.UUID,
			ContentURI:       complementaryContentURI,
			IsMarkedDeleted:  markDelete,
			PublishReference: tid,
			LastModified:     lmd,
		},
		Type:   contentType,
		Brands: model.BuildBrands(),
	}
}

func (ccm *ComplementaryContentCPHMapper) buildCCAlternativeTitles(promoTitle string) *model.AlternativeTitles {
	promoTitle = strings.TrimSpace(promoTitle)

	if promoTitle == "" {
		return nil
	}
	return &model.AlternativeTitles{PromotionalTitle: promoTitle}
}

func (ccm *ComplementaryContentCPHMapper) buildCCAlternativeImages(fileRef string) *model.AlternativeImages {
	if fileRef == "" {
		return nil
	}
	imageUUID := extractImageUUID(fileRef)
	return &model.AlternativeImages{PromotionalImage: &model.PromotionalImage{Id: fmt.Sprintf(ccm.apiHostFormat, imageUUID)}}
}

func extractImageUUID(fileRef string) string {
	return strings.Split(fileRef, "uuid=")[1]
}

func (ccm *ComplementaryContentCPHMapper) buildCCAlternativeStandfirsts(promoStandfirst string) *model.AlternativeStandfirsts {
	promoStandfirst = strings.TrimSpace(promoStandfirst)
	if promoStandfirst == "" {
		return nil
	}
	return &model.AlternativeStandfirsts{PromotionalStandfirst: promoStandfirst}
}

func (ccm *ComplementaryContentCPHMapper) setBrands(uuid, tid string, cc *model.UppComplementaryContent) error {
	content, err := ccm.client.GetContent(uuid, tid)
	if err != nil {
		return fmt.Errorf("failed to get content from document-store-api: %v", err.Error())
	}
	cc.Brands = content.Brands
	return nil
}
