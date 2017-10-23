package model

type UppComplementaryContent struct {
	UppCoreContent
	AlternativeTitles      *AlternativeTitles      `json:"alternativeTitles"`
	AlternativeImages      *AlternativeImages      `json:"alternativeImages"`
	AlternativeStandfirsts *AlternativeStandfirsts `json:"alternativeStandfirsts"`
	Brands                 []Brand                 `json:"brands"`
	Type                   string                  `json:"type"`
}

type AlternativeImages struct {
	PromotionalImage *PromotionalImage `json:"promotionalImage"`
}

type PromotionalImage struct {
	Id string `json:"id"`
}

type AlternativeStandfirsts struct {
	PromotionalStandfirst string `json:"promotionalStandfirst"`
}
