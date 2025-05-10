package repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/redis/go-redis/v9"
)

type RedisLiveRepo struct {
	redis               *redis.Client
	ctx                 context.Context
	dataKeyPrefix       string
	channelKeyPrefix    string
	createSessionScript *redis.Script
}

func NewRedisLiveRepo(redisClient *redis.Client) *RedisLiveRepo {
	ctx := context.Background()

	repo := &RedisLiveRepo{
		redis:            redisClient,
		ctx:              ctx,
		dataKeyPrefix:    "live_session_data:",
		channelKeyPrefix: "live_session_channel:",
		createSessionScript: redis.NewScript(`
local prefix = ARGV[1]
local field_data = {}

for i = 2, #ARGV do
    table.insert(field_data, ARGV[i])
end

for i = 1, 100 do
    local n = math.random(0, 9999)
    local suffix = string.format("%04d", n)
    local candidate = prefix .. suffix
    if redis.call('EXISTS', candidate) == 0 then
        redis.call('HSET', candidate, unpack(field_data))
        return suffix
    end
end

return nil
`),
	}

	return repo
}

func (r *RedisLiveRepo) dataKey(key string) string {
	return r.dataKeyPrefix + key
}

func (r *RedisLiveRepo) channelKey(key string) string {
	return r.channelKeyPrefix + key
}

func (r *RedisLiveRepo) CleanUp(minUpdatedAt time.Time) ([]string, error) {
	filenames := make([]string, 0)
	minUpdatedAtUnix := minUpdatedAt.Unix()

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = r.redis.Scan(r.ctx, cursor, r.dataKeyPrefix+"*", 10).Result()
		if err != nil {
			return nil, err
		}

		for _, dataKey := range keys {
			key := strings.TrimPrefix(dataKey, r.dataKeyPrefix)

			data, err := r.redis.HMGet(r.ctx, dataKey, "updatedAt", "fileName").Result()
			if err != nil {
				continue
			}

			updatedAtUnix, err := strconv.ParseInt(data[0].(string), 10, 64)
			if err != nil {
				continue
			}

			fileName, ok := data[1].(string)
			if !ok {
				continue
			}

			if updatedAtUnix >= minUpdatedAtUnix {
				continue
			}

			channelKey := r.channelKey(key)
			subscribersCount, err := r.redis.PubSubNumSub(r.ctx, channelKey).Result()
			if err != nil {
				continue
			}

			if subscribersCount[channelKey] > 0 {
				continue
			}

			r.redis.Del(r.ctx, dataKey)

			filenames = append(filenames, fileName)
		}

		if cursor == 0 {
			break
		}
	}

	return filenames, nil
}

func (r *RedisLiveRepo) toHash(session *models.LiveSession) map[string]string {
	return map[string]string{
		"url":         session.URL,
		"currentPage": strconv.Itoa(session.CurrentPage),
		"token":       session.Token,
		"fileName":    session.FileName,
		"updatedAt":   strconv.FormatInt(session.UpdatedAt.Unix(), 10),
	}
}

func (r *RedisLiveRepo) toPairs(session *models.LiveSession) []interface{} {
	hash := r.toHash(session)
	pairs := make([]interface{}, 0, len(hash)*2)
	for k, v := range hash {
		pairs = append(pairs, k, v)
	}
	return pairs
}

func (r *RedisLiveRepo) fromHash(hash map[string]string) (*models.LiveSession, error) {
	currentPage, err := strconv.Atoi(hash["currentPage"])
	if err != nil {
		return nil, err
	}

	updatedAt, err := strconv.ParseInt(hash["updatedAt"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &models.LiveSession{
		URL:         hash["url"],
		CurrentPage: currentPage,
		Token:       hash["token"],
		FileName:    hash["fileName"],
		UpdatedAt:   time.Unix(updatedAt, 0),
	}, nil
}

func (r *RedisLiveRepo) GetSession(key string) (*models.LiveSession, error) {
	hash, err := r.redis.HGetAll(r.ctx, r.dataKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return r.fromHash(hash)
}

func (r *RedisLiveRepo) CreateSession(session *models.LiveSession) (string, error) {
	data := append([]interface{}{r.dataKeyPrefix}, r.toPairs(session)...)
	key, err := r.createSessionScript.Run(r.ctx, r.redis, []string{}, data...).Result()
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		return "", err
	}

	if keyString, ok := key.(string); ok {
		return keyString, nil
	}

	return "", errors.New("invalid value returned from redis")
}

func (r *RedisLiveRepo) UpdateSession(key string, session *models.LiveSession) error {
	return r.redis.HSet(r.ctx, r.dataKey(key), r.toHash(session)).Err()
}

func (r *RedisLiveRepo) ChangeSessionPage(key string, currentPage int) error {
	return r.redis.HSet(r.ctx, r.dataKey(key),
		"currentPage", currentPage,
	).Err()
}

func (r *RedisLiveRepo) DeleteSession(key string) error {
	return r.redis.Del(r.ctx, r.dataKey(key)).Err()
}

func (r *RedisLiveRepo) PublishEvent(key string, event *dtos.Event) error {
	r.redis.HSet(r.ctx, r.dataKey(key), "updatedAt", time.Now().Unix())

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.redis.Publish(r.ctx, r.channelKey(key), data).Err()
}

func (r *RedisLiveRepo) Subscribe(key string) (chan *dtos.Event, func(), error) {
	pubsub := r.redis.Subscribe(r.ctx, r.channelKey(key))
	ch := make(chan *dtos.Event)

	go func() {
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			var event dtos.Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			ch <- &event
		}
	}()

	return ch, func() {
		pubsub.Close()
		close(ch)
	}, nil
}
