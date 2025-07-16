package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/redis/go-redis/v9"
)

type RedisLiturgyRepo struct {
	redis *redis.Client
}

func (r *RedisLiturgyRepo) getKey(date string) string {
	return fmt.Sprintf("liturgy:%s", date)
}

func (r *RedisLiturgyRepo) GetDay(date string) (dtos.LiturgyItems, bool) {
	liturgy := dtos.LiturgyItems{}
	items, err := r.redis.HGetAll(context.Background(), r.getKey(date)).Result()
	if err != nil || items == nil || len(items) == 0 {
		return liturgy, false
	}

	liturgy.Psalm = items["psalm"]
	liturgy.Acclamation = items["acclamation"]
	liturgy.AcclamationVerse = items["acclamationVerse"]

	return liturgy, true
}

func (r *RedisLiturgyRepo) StoreDay(date string, liturgy dtos.LiturgyItems) error {
	key := r.getKey(date)
	err := r.redis.HSet(context.Background(), key,
		"psalm", liturgy.Psalm,
		"acclamation", liturgy.Acclamation,
		"acclamationVerse", liturgy.AcclamationVerse,
	).Err()
	if err != nil {
		return err
	}

	r.redis.ExpireAt(context.Background(), key, r.getExpirationTime(date)).Err()

	return nil
}

func (r *RedisLiturgyRepo) getExpirationTime(date string) time.Time {
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.Local)

	if err != nil {
		return time.Now().Add(24 * time.Hour)
	}

	return parsedDate.Add(24 * time.Hour)
}

func NewRedisLiturgyRepo(redis *redis.Client) *RedisLiturgyRepo {
	return &RedisLiturgyRepo{
		redis: redis,
	}
}
