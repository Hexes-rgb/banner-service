package bannerservice

import (
	"banner-service/internal/models"
	"banner-service/internal/utils"
	"context"
	"errors"
	"time"
)

type CacheBannerRepository interface {
	GetBanner(ctx context.Context, key string) (*models.Banner, error)
	SetBanner(ctx context.Context, key string, banner *models.Banner, ttl time.Duration) error
}

type DBBannerRepository interface {
	GetBanner(ctx context.Context, featureID, tagID int, isAdmin bool) (*models.Banner, error)
	GetBanners(ctx context.Context, featureID, tagID, limit, offset int) ([]*models.Banner, error)
	CreateBanner(ctx context.Context, banner *models.Banner) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, banner *models.Banner) error
	DeleteBanner(ctx context.Context, bannerID int) error
}

type BannerService struct {
	cacheRepo CacheBannerRepository
	dbRepo    DBBannerRepository
}

func NewBannerService(cacheRepo CacheBannerRepository, dbRepo DBBannerRepository) *BannerService {
	return &BannerService{
		cacheRepo: cacheRepo,
		dbRepo:    dbRepo,
	}
}

func (s *BannerService) GetBanner(ctx context.Context, tagID, featureID int, useLastRevision, isAdmin bool) (*models.Banner, error) {
	if !useLastRevision {
		cachedBanner, err := s.cacheRepo.GetBanner(ctx, utils.MakeCacheKey(featureID, tagID))
		if err != nil {
			return nil, err
		}
		if cachedBanner != nil {
			return cachedBanner, nil
		}
	}

	dbBanner, err := s.dbRepo.GetBanner(ctx, featureID, tagID, isAdmin)
	if err != nil {
		return nil, err
	}

	if dbBanner != nil {
		_ = s.cacheRepo.SetBanner(ctx, utils.MakeCacheKey(featureID, tagID), dbBanner, time.Duration(time.Minute*5))
	}

	return dbBanner, nil
}

func (s *BannerService) GetBanners(ctx context.Context, featureID, tagID, limit, offset int) ([]*models.Banner, error) {
	if featureID == -1 {
		featureID = 0
	}
	if tagID == -1 {
		tagID = 0
	}

	if limit <= 0 {
		limit = 0
	}

	if offset < 0 {
		offset = 0
	}

	banners, err := s.dbRepo.GetBanners(ctx, featureID, tagID, limit, offset)
	if err != nil {
		return nil, err
	}

	return banners, nil
}

func (s *BannerService) CreateBanner(ctx context.Context, banner *models.Banner) (int, error) {
	if banner == nil {
		return 0, errors.New("banner не может быть nil")
	}
	if len(banner.TagIDs) == 0 {
		return 0, errors.New("должен быть указан хотя бы один tag_id")
	}
	if banner.FeatureID == 0 {
		return 0, errors.New("неверный feature_id")
	}
	if banner.Content == nil {
		return 0, errors.New("неверное содержимое баннера")
	}

	bannerID, err := s.dbRepo.CreateBanner(ctx, banner)
	if err != nil {
		return 0, err
	}

	return bannerID, nil
}

func (s *BannerService) UpdateBanner(ctx context.Context, bannerID int, banner *models.Banner) error {
	if banner == nil {
		return errors.New("banner не может быть nil")
	}
	if len(banner.TagIDs) == 0 {
		return errors.New("должен быть указан хотя бы один tag_id")
	}
	if banner.FeatureID == 0 {
		return errors.New("неверный feature_id")
	}
	if banner.Content == nil {
		return errors.New("неверное содержимое баннера")
	}

	err := s.dbRepo.UpdateBanner(ctx, bannerID, banner)
	if err != nil {
		return err
	}

	return nil
}

func (s *BannerService) DeleteBanner(ctx context.Context, bannerID int) error {
	return s.dbRepo.DeleteBanner(ctx, bannerID)
}
