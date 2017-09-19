package model

type UppComplementaryContent struct {
	UppCoreContent
	AlternativeTitles      *AlternativeTitles      `json:"alternativeTitles"`
	AlternativeImages      *AlternativeImages      `json:"alternativeImages"`
	AlternativeStandfirsts *AlternativeStandfirsts `json:"alternativeStandfirsts"`
}

type AlternativeImages struct {
	PromotionalImage string `json:"promotionalImage"`
}

type AlternativeStandfirsts struct {
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}
