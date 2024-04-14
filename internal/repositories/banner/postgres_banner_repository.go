package bannerrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"banner-service/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresBannerRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBannerRepository(pool *pgxpool.Pool) *PostgresBannerRepository {
	return &PostgresBannerRepository{
		pool: pool,
	}
}

func (r *PostgresBannerRepository) GetBanner(ctx context.Context, featureID, tagID int, isAdmin bool) (*models.Banner, error) {
	whereConditions := "WHERE b.feature_id = $1 AND bt.tag_id = $2"
	if !isAdmin {
		whereConditions += " AND b.is_active = TRUE"
	}

	query := fmt.Sprintf(`
    SELECT b.banner_id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
    FROM banners b
    INNER JOIN banner_tag bt ON b.banner_id = bt.banner_id
    %s
    GROUP BY b.banner_id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
    LIMIT 1
    `, whereConditions)

	banner := &models.Banner{}

	if err := r.pool.QueryRow(ctx, query, featureID, tagID).Scan(
		&banner.BannerID,
		&banner.FeatureID,
		&banner.Content,
		&banner.IsActive,
		&banner.CreatedAt,
		&banner.UpdatedAt,
	); err != nil {
		return nil, err
	}

	tagQuery := `
	SELECT tag_id
	FROM banner_tag
	WHERE banner_id = $1
	`
	tagRows, err := r.pool.Query(ctx, tagQuery, banner.BannerID)
	if err != nil {
		return nil, err
	}
	var tagIDs []int
	for tagRows.Next() {
		var tagID int
		if err := tagRows.Scan(&tagID); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tagID)
	}
	if err := tagRows.Err(); err != nil {
		return nil, err
	}
	tagRows.Close()

	banner.TagIDs = tagIDs

	return banner, nil
}

func (r *PostgresBannerRepository) GetBanners(ctx context.Context, featureID, tagID, limit, offset int) ([]*models.Banner, error) {
	var queryParams []interface{}
	baseQuery := `
	SELECT b.banner_id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
	FROM banners b
	`

	whereConditions := []string{"1=1"}
	if featureID > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("b.feature_id = $%d", len(queryParams)+1))
		queryParams = append(queryParams, featureID)
	}
	if tagID > 0 {
		baseQuery = `
		SELECT b.banner_id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
		FROM banners b
		LEFT JOIN banner_tag bt ON b.banner_id = bt.banner_id
		`
		whereConditions = append(whereConditions, fmt.Sprintf("bt.tag_id = $%d", len(queryParams)+1))
		queryParams = append(queryParams, tagID)
	}

	if len(whereConditions) > 0 {
		baseQuery += "WHERE " + strings.Join(whereConditions, " AND ")
	}

	baseQuery += `
	GROUP BY b.banner_id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
	`
	var query string
	if limit > 0 && offset >= 0 {
		query = fmt.Sprintf("%s ORDER BY b.updated_at DESC LIMIT $%d OFFSET $%d", baseQuery, len(queryParams)+1, len(queryParams)+2)
		queryParams = append(queryParams, limit, offset)
	} else {
		query = fmt.Sprintf("%s ORDER BY b.updated_at DESC", baseQuery)
	}

	rows, err := r.pool.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	banners := make([]*models.Banner, 0)
	for rows.Next() {
		banner := &models.Banner{}
		if err := rows.Scan(
			&banner.BannerID,
			&banner.FeatureID,
			&banner.Content,
			&banner.IsActive,
			&banner.CreatedAt,
			&banner.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tagQuery := `
		SELECT tag_id
		FROM banner_tag
		WHERE banner_id = $1
		`
		tagRows, err := r.pool.Query(ctx, tagQuery, banner.BannerID)
		if err != nil {
			return nil, err
		}
		var tagIDs []int
		for tagRows.Next() {
			var tagID int
			if err := tagRows.Scan(&tagID); err != nil {
				return nil, err
			}
			tagIDs = append(tagIDs, tagID)
		}
		if err := tagRows.Err(); err != nil {
			return nil, err
		}
		tagRows.Close()

		banner.TagIDs = tagIDs
		banners = append(banners, banner)
	}

	return banners, nil
}

func (r *PostgresBannerRepository) CreateBanner(ctx context.Context, banner *models.Banner) (int, error) {
	contentJSON, err := json.Marshal(banner.Content)
	if err != nil {
		return 0, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO banners (feature_id, content, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING banner_id
	`

	var bannerID int
	if err := tx.QueryRow(ctx, query, banner.FeatureID, contentJSON, banner.IsActive, time.Now(), time.Now()).Scan(&bannerID); err != nil {
		return 0, err
	}

	for _, tagID := range banner.TagIDs {
		_, err = tx.Exec(ctx, "INSERT INTO banner_tag (banner_id, tag_id) VALUES ($1, $2)", bannerID, tagID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return bannerID, nil
}

func (r *PostgresBannerRepository) UpdateBanner(ctx context.Context, bannerID int, banner *models.Banner) error {
	contentJSON, err := json.Marshal(banner.Content)
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := `
	UPDATE banners
	SET feature_id = $1, content = $2, is_active = $3, updated_at = $4
	WHERE banner_id = $5
	`

	if cmdTag, err := tx.Exec(ctx, query, banner.FeatureID, contentJSON, banner.IsActive, time.Now(), bannerID); err != nil {
		return err
	} else if cmdTag.RowsAffected() != 1 {
		return errors.New("no rows affected")
	}

	if _, err = tx.Exec(ctx, "DELETE FROM banner_tag WHERE banner_id = $1", bannerID); err != nil {
		return err
	}

	for _, tagID := range banner.TagIDs {
		if _, err = tx.Exec(ctx, "INSERT INTO banner_tag (banner_id, tag_id) VALUES ($1, $2)", bannerID, tagID); err != nil {
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *PostgresBannerRepository) DeleteBanner(ctx context.Context, bannerID int) error {
	var ErrNoRowsAffected = errors.New("no rows affected")
	query := `
	DELETE FROM banners
	WHERE banner_id = $1
	`

	if cmdTag, err := r.pool.Exec(ctx, query, bannerID); err != nil {
		return err
	} else if cmdTag.RowsAffected() != 1 {
		return ErrNoRowsAffected
	}

	return nil
}
