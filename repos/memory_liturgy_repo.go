package repos

import "github.com/hejmsdz/goslides/dtos"

type MemoryLiturgyRepo struct {
	days map[string]dtos.LiturgyItems
}

func (r *MemoryLiturgyRepo) GetDay(date string) (dtos.LiturgyItems, bool) {
	liturgy, ok := r.days[date]
	return liturgy, ok
}

func (r *MemoryLiturgyRepo) StoreDay(date string, liturgy dtos.LiturgyItems) error {
	r.days[date] = liturgy
	return nil
}

func NewMemoryLiturgyRepo() *MemoryLiturgyRepo {
	return &MemoryLiturgyRepo{
		days: make(map[string]dtos.LiturgyItems),
	}
}
