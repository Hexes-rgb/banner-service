package banner

import (
	"banner-service/internal/models"
	"context"
)

type DBBannerRepository interface {
	GetBanner(ctx context.Context, tagID, featureID int, isAdmin bool) (*models.Banner, error)
	GetBanners(ctx context.Context, featureID, tagID, limit, offset int) ([]*models.Banner, error)
	CreateBanner(ctx context.Context, banner *models.Banner) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, banner *models.Banner) error
	DeleteBanner(ctx context.Context, bannerID int) error
}
