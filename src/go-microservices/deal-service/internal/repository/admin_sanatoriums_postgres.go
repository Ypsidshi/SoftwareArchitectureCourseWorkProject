package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"coursework/deal-service/internal/domain"
	"github.com/google/uuid"
)

var ErrSanatoriumHasBookings = errors.New("sanatorium has active bookings")

type NewSanatorium struct {
	Name            string
	Description     string
	City            string
	Address         string
	DistanceToSeaKM float64
	Amenities       []string
	ImageURLs       []string
	PricePerNight   float64
	TotalPlaces     int
	Latitude         *float64
	Longitude        *float64
	MedicalProfiles  []string
}

type UpdateSanatorium struct {
	ID              string
	Name            string
	Description     string
	City            string
	Address         string
	DistanceToSeaKM float64
	Amenities       []string
	ImageURLs       []string
	PricePerNight   float64
	TotalPlaces     int
	Latitude         *float64
	Longitude        *float64
	MedicalProfiles  []string
}

func encodeStringSlice(items []string) []byte {
	if items == nil {
		items = []string{}
	}
	raw, _ := json.Marshal(items)
	return raw
}

func (r *Repository) CreateSanatorium(ctx context.Context, in NewSanatorium) (domain.Sanatorium, error) {
	const q = `
INSERT INTO deal.sanatoriums (
	id, name, description, city, address, distance_to_sea_km, amenities, image_urls,
	price_per_night, total_places, latitude, longitude
)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11, $12)
RETURNING id, name, description, city, address, distance_to_sea_km, amenities, image_urls,
          price_per_night, total_places, latitude, longitude, created_at, updated_at`

	id := uuid.NewString()
	row := r.db.QueryRowContext(ctx, q,
		id, in.Name, in.Description, in.City, in.Address, in.DistanceToSeaKM,
		encodeStringSlice(in.Amenities), encodeStringSlice(in.ImageURLs),
		in.PricePerNight, in.TotalPlaces, in.Latitude, in.Longitude,
	)
	item, err := scanSanatoriumBase(row)
	if err != nil {
		return domain.Sanatorium{}, err
	}
	if err := r.syncMedicalProfiles(ctx, id, in.MedicalProfiles); err != nil {
		return domain.Sanatorium{}, err
	}
	profiles, _ := r.loadMedicalProfiles(ctx, id)
	item.MedicalProfiles = profiles
	return item, nil
}

func (r *Repository) UpdateSanatorium(ctx context.Context, in UpdateSanatorium) (domain.Sanatorium, error) {
	const q = `
UPDATE deal.sanatoriums
SET name = $2, description = $3, city = $4, address = $5, distance_to_sea_km = $6,
    amenities = $7::jsonb, image_urls = $8::jsonb, price_per_night = $9, total_places = $10,
    latitude = $11, longitude = $12, updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, city, address, distance_to_sea_km, amenities, image_urls,
          price_per_night, total_places, latitude, longitude, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, q,
		in.ID, in.Name, in.Description, in.City, in.Address, in.DistanceToSeaKM,
		encodeStringSlice(in.Amenities), encodeStringSlice(in.ImageURLs),
		in.PricePerNight, in.TotalPlaces, in.Latitude, in.Longitude,
	)
	item, err := scanSanatoriumBase(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Sanatorium{}, ErrSanatoriumNotFound
		}
		return domain.Sanatorium{}, err
	}
	if err := r.syncMedicalProfiles(ctx, in.ID, in.MedicalProfiles); err != nil {
		return domain.Sanatorium{}, err
	}
	profiles, _ := r.loadMedicalProfiles(ctx, in.ID)
	item.MedicalProfiles = profiles
	return item, nil
}

func (r *Repository) syncMedicalProfiles(ctx context.Context, sanatoriumID string, names []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM deal.sanatorium_medical_profiles WHERE sanatorium_id = $1`, sanatoriumID); err != nil {
		return err
	}
	for _, raw := range names {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		var profileID string
		err := r.db.QueryRowContext(ctx, `SELECT id FROM deal.medical_profiles WHERE LOWER(name) = $1`, name).Scan(&profileID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		if _, err := r.db.ExecContext(ctx, `
INSERT INTO deal.sanatorium_medical_profiles (sanatorium_id, profile_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING`, sanatoriumID, profileID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListMedicalProfileNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM deal.medical_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func (r *Repository) DeleteSanatorium(ctx context.Context, id string) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM deal.bookings WHERE sanatorium_id = $1 AND status = 'confirmed'`, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrSanatoriumHasBookings
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM deal.sanatoriums WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrSanatoriumNotFound
	}
	return nil
}

func (r *Repository) loadMedicalProfiles(ctx context.Context, sanatoriumID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT mp.name
FROM deal.sanatorium_medical_profiles smp
JOIN deal.medical_profiles mp ON mp.id = smp.profile_id
WHERE smp.sanatorium_id = $1
ORDER BY mp.name`, sanatoriumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func (r *Repository) ListSanatoriumsAdmin(ctx context.Context, page, pageSize int) ([]domain.Sanatorium, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deal.sanatoriums`).Scan(&total); err != nil {
		return nil, 0, err
	}
	const listQuery = `
SELECT
	s.id, s.name, s.description, s.city, s.address, s.distance_to_sea_km,
	s.amenities, s.image_urls, s.price_per_night, s.total_places,
	COALESCE((
		SELECT json_agg(mp.name ORDER BY mp.name)
		FROM deal.sanatorium_medical_profiles smp
		JOIN deal.medical_profiles mp ON mp.id = smp.profile_id
		WHERE smp.sanatorium_id = s.id
	), '[]'::json) AS medical_profiles,
	s.latitude, s.longitude, s.created_at, s.updated_at
FROM deal.sanatoriums s
ORDER BY s.name
LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, listQuery, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.Sanatorium, 0, pageSize)
	for rows.Next() {
		item, err := scanSanatorium(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
