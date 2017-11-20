package model

type DocStoreUppContent struct {
	UppCoreContent
	Brands []Brand `json:"brands"`
}
