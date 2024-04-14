package bannerrepo

import (
	"banner-service/internal/models"
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisBannerRepository struct {
	client *redis.Client
}

func NewRedisBannerRepository(client *redis.Client) *RedisBannerRepository {
	return &RedisBannerRepository{client: client}
}

func (r *RedisBannerRepository) GetBanner(ctx context.Context, key string) (*models.Banner, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var banner models.Banner
	err = json.Unmarshal([]byte(val), &banner)
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (r *RedisBannerRepository) SetBanner(ctx context.Context, key string, banner *models.Banner, ttl time.Duration) error {
	val, err := json.Marshal(banner)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, val, ttl).Err()
}
