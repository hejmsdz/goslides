package repos

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisNonceRepo struct {
	redis *redis.Client
	ctx   context.Context
}

func NewRedisNonceRepo(redis *redis.Client) *RedisNonceRepo {
	return &RedisNonceRepo{redis: redis, ctx: context.Background()}
}

func (r *RedisNonceRepo) getKey(token string) string {
	return "nonce:" + token
}

func (r *RedisNonceRepo) GetUserIdFromNonce(token string) (uint, error) {
	val := r.redis.GetDel(r.ctx, r.getKey(token)).Val()
	if val == "" {
		return 0, errors.New("invalid nonce")
	}

	id, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func (r *RedisNonceRepo) CreateNonce(token string, userID uint, expiration time.Duration) error {
	return r.redis.Set(r.ctx, r.getKey(token), userID, expiration).Err()
}
