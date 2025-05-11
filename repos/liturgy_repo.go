package repos

import "github.com/hejmsdz/goslides/dtos"

type LiturgyRepo interface {
	GetDay(date string) (dtos.LiturgyItems, bool)
	StoreDay(date string, liturgy dtos.LiturgyItems) error
}
