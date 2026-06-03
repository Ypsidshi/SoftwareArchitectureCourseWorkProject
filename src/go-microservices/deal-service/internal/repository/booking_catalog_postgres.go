package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"coursework/deal-service/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrBookingNotFound        = errors.New("booking not found")
	ErrSanatoriumNotFound     = errors.New("sanatorium not found")
	ErrSanatoriumNotAvailable = errors.New("sanatorium is not available for selected dates")
	ErrGuestsExceedCapacity   = errors.New("guests exceed sanatorium capacity")
)

type SanatoriumFilter struct {
	Page               int
	PageSize           int
	City               string
	ProfileNames       []string
	MaxDistanceToSeaKM *float64
	PriceMin           *float64
	PriceMax           *float64
	CheckIn            *time.Time
	CheckOut           *time.Time
	Sort               string
}

type NewBooking struct {
	ClientID     string
	SanatoriumID string
	CheckIn      time.Time
	CheckOut     time.Time
	Guests       int
}

type UpdateBooking struct {
	ID       string
	ClientID string
	CheckIn  time.Time
	CheckOut time.Time
	Guests   int
}

func (r *Repository) ListSanatoriums(ctx context.Context, filter SanatoriumFilter) ([]domain.Sanatorium, int, error) {
	whereSQL, args := buildSanatoriumWhere(filter)

	countQuery := `SELECT COUNT(*) FROM deal.sanatoriums s WHERE ` + whereSQL
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "s.created_at DESC"
	switch filter.Sort {
	case "price_asc":
		orderBy = "s.price_per_night ASC"
	case "price_desc":
		orderBy = "s.price_per_night DESC"
	case "distance_asc":
		orderBy = "s.distance_to_sea_km ASC"
	case "distance_desc":
		orderBy = "s.distance_to_sea_km DESC"
	}

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	listQuery := `
SELECT
	s.id,
	s.name,
	s.description,
	s.city,
	s.address,
	s.distance_to_sea_km,
	s.amenities,
	s.image_urls,
	s.price_per_night,
	s.total_places,
	COALESCE((
		SELECT json_agg(mp.name ORDER BY mp.name)
		FROM deal.sanatorium_medical_profiles smp
		JOIN deal.medical_profiles mp ON mp.id = smp.profile_id
		WHERE smp.sanatorium_id = s.id
	), '[]'::json) AS medical_profiles,
	s.latitude,
	s.longitude,
	s.created_at,
	s.updated_at
FROM deal.sanatoriums s
WHERE ` + whereSQL + `
ORDER BY ` + orderBy + `
LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Sanatorium, 0, filter.PageSize)
	for rows.Next() {
		item, scanErr := scanSanatorium(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) GetSanatoriumByID(ctx context.Context, id string) (domain.Sanatorium, error) {
	const query = `
SELECT
	s.id,
	s.name,
	s.description,
	s.city,
	s.address,
	s.distance_to_sea_km,
	s.amenities,
	s.image_urls,
	s.price_per_night,
	s.total_places,
	COALESCE((
		SELECT json_agg(mp.name ORDER BY mp.name)
		FROM deal.sanatorium_medical_profiles smp
		JOIN deal.medical_profiles mp ON mp.id = smp.profile_id
		WHERE smp.sanatorium_id = s.id
	), '[]'::json) AS medical_profiles,
	s.latitude,
	s.longitude,
	s.created_at,
	s.updated_at
FROM deal.sanatoriums s
WHERE s.id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanSanatorium(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Sanatorium{}, ErrSanatoriumNotFound
		}
		return domain.Sanatorium{}, err
	}
	return item, nil
}

func (r *Repository) CreateBooking(ctx context.Context, in NewBooking) (domain.Booking, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1));`, in.SanatoriumID); err != nil {
		return domain.Booking{}, err
	}

	totalPlaces, err := getSanatoriumCapacity(ctx, tx, in.SanatoriumID)
	if err != nil {
		return domain.Booking{}, err
	}
	if in.Guests > totalPlaces {
		return domain.Booking{}, ErrGuestsExceedCapacity
	}

	available, err := isSanatoriumAvailable(ctx, tx, in.SanatoriumID, in.CheckIn, in.CheckOut, in.Guests, nil)
	if err != nil {
		return domain.Booking{}, err
	}
	if !available {
		return domain.Booking{}, ErrSanatoriumNotAvailable
	}

	const insertQuery = `
INSERT INTO deal.bookings (id, client_id, sanatorium_id, check_in, check_out, guests, status)
VALUES ($1, $2, $3, $4, $5, $6, 'confirmed')
RETURNING ` + bookingRowColumns

	id := uuid.NewString()
	var booking domain.Booking
	row := tx.QueryRowContext(ctx, insertQuery,
		id, in.ClientID, in.SanatoriumID, in.CheckIn, in.CheckOut, in.Guests,
	)
	if err := scanBooking(row, &booking); err != nil {
		return domain.Booking{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) UpdateBooking(ctx context.Context, in UpdateBooking) (domain.Booking, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback()

	var sanatoriumID, status string
	err = tx.QueryRowContext(ctx, `
SELECT sanatorium_id, status
FROM deal.bookings
WHERE id = $1 AND client_id = $2
FOR UPDATE`,
		in.ID, in.ClientID,
	).Scan(&sanatoriumID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	if status == "cancelled" {
		return domain.Booking{}, ErrBookingNotFound
	}

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1));`, sanatoriumID); err != nil {
		return domain.Booking{}, err
	}

	totalPlaces, err := getSanatoriumCapacity(ctx, tx, sanatoriumID)
	if err != nil {
		return domain.Booking{}, err
	}
	if in.Guests > totalPlaces {
		return domain.Booking{}, ErrGuestsExceedCapacity
	}

	available, err := isSanatoriumAvailable(ctx, tx, sanatoriumID, in.CheckIn, in.CheckOut, in.Guests, &in.ID)
	if err != nil {
		return domain.Booking{}, err
	}
	if !available {
		return domain.Booking{}, ErrSanatoriumNotAvailable
	}

	const updateQuery = `
UPDATE deal.bookings
SET check_in = $3, check_out = $4, guests = $5, updated_at = NOW()
WHERE id = $1 AND client_id = $2
RETURNING ` + bookingRowColumns

	var booking domain.Booking
	row := tx.QueryRowContext(ctx, updateQuery, in.ID, in.ClientID, in.CheckIn, in.CheckOut, in.Guests)
	if err := scanBooking(row, &booking); err != nil {
		return domain.Booking{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) CancelBooking(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	const query = `
UPDATE deal.bookings
SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
WHERE id = $1 AND client_id = $2 AND status <> 'cancelled'
RETURNING ` + bookingRowColumns

	var booking domain.Booking
	err := scanBooking(r.db.QueryRowContext(ctx, query, bookingID, clientID), &booking)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) GetBookingByID(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	return r.GetBookingForClient(ctx, bookingID, clientID)
}

func (r *Repository) ListBookingsByClient(ctx context.Context, clientID string, page, pageSize int) ([]domain.Booking, int, error) {
	const countQuery = `SELECT COUNT(*) FROM deal.bookings WHERE client_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, clientID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQuery = `
SELECT ` + bookingSelectColumns + `
FROM deal.bookings b
WHERE b.client_id = $1
ORDER BY b.created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, listQuery, clientID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Booking, 0, pageSize)
	for rows.Next() {
		var booking domain.Booking
		if err := scanBooking(rows, &booking); err != nil {
			return nil, 0, err
		}
		items = append(items, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) CheckAvailability(ctx context.Context, sanatoriumID string, checkIn, checkOut time.Time, excludeBookingID *string) (bool, error) {
	return isSanatoriumAvailable(ctx, r.db, sanatoriumID, checkIn, checkOut, 1, excludeBookingID)
}

func buildSanatoriumWhere(filter SanatoriumFilter) (string, []any) {
	conditions := []string{"1=1"}
	args := make([]any, 0, 8)

	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if city := strings.TrimSpace(filter.City); city != "" {
		p := addArg(strings.ToLower(city))
		conditions = append(conditions, "LOWER(s.city) = "+p)
	}

	if filter.MaxDistanceToSeaKM != nil {
		p := addArg(*filter.MaxDistanceToSeaKM)
		conditions = append(conditions, "s.distance_to_sea_km <= "+p)
	}

	if filter.PriceMin != nil {
		p := addArg(*filter.PriceMin)
		conditions = append(conditions, "s.price_per_night >= "+p)
	}

	if filter.PriceMax != nil {
		p := addArg(*filter.PriceMax)
		conditions = append(conditions, "s.price_per_night <= "+p)
	}

	profiles := make([]string, 0, len(filter.ProfileNames))
	for _, profile := range filter.ProfileNames {
		trimmed := strings.ToLower(strings.TrimSpace(profile))
		if trimmed != "" {
			profiles = append(profiles, trimmed)
		}
	}
	if len(profiles) > 0 {
		inParts := make([]string, 0, len(profiles))
		for _, profile := range profiles {
			inParts = append(inParts, addArg(profile))
		}
		conditions = append(conditions, `
EXISTS (
	SELECT 1
	FROM deal.sanatorium_medical_profiles smp
	JOIN deal.medical_profiles mp ON mp.id = smp.profile_id
	WHERE smp.sanatorium_id = s.id
	  AND LOWER(mp.name) IN (`+strings.Join(inParts, ",")+`)
)`)
	}

	if filter.CheckIn != nil && filter.CheckOut != nil {
		checkInParam := addArg(*filter.CheckIn)
		checkOutParam := addArg(*filter.CheckOut)
		conditions = append(conditions, `
	COALESCE((
	SELECT SUM(b.guests)
	FROM deal.bookings b
	WHERE b.sanatorium_id = s.id
	  AND b.status = 'confirmed'
	  AND daterange(b.check_in, b.check_out, '[)')
	      && daterange(`+checkInParam+`::date, `+checkOutParam+`::date, '[)')
), 0) < s.total_places`)
	}

	return strings.Join(conditions, " AND "), args
}

type sanatoriumScanner interface {
	Scan(dest ...any) error
}

func scanSanatorium(scanner sanatoriumScanner) (domain.Sanatorium, error) {
	var item domain.Sanatorium
	var amenitiesRaw, imagesRaw, profilesRaw []byte
	var latitude, longitude sql.NullFloat64
	err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.City,
		&item.Address,
		&item.DistanceToSeaKM,
		&amenitiesRaw,
		&imagesRaw,
		&item.PricePerNight,
		&item.TotalPlaces,
		&profilesRaw,
		&latitude,
		&longitude,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Sanatorium{}, err
	}

	item.Amenities = decodeStringSlice(amenitiesRaw)
	item.ImageURLs = decodeStringSlice(imagesRaw)
	item.MedicalProfiles = decodeStringSlice(profilesRaw)
	if latitude.Valid {
		item.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		item.Longitude = &longitude.Float64
	}
	return item, nil
}

func scanSanatoriumBase(scanner sanatoriumScanner) (domain.Sanatorium, error) {
	var item domain.Sanatorium
	var amenitiesRaw, imagesRaw []byte
	var latitude, longitude sql.NullFloat64
	err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.City,
		&item.Address,
		&item.DistanceToSeaKM,
		&amenitiesRaw,
		&imagesRaw,
		&item.PricePerNight,
		&item.TotalPlaces,
		&latitude,
		&longitude,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Sanatorium{}, err
	}

	item.Amenities = decodeStringSlice(amenitiesRaw)
	item.ImageURLs = decodeStringSlice(imagesRaw)
	item.MedicalProfiles = []string{}
	if latitude.Valid {
		item.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		item.Longitude = &longitude.Float64
	}
	return item, nil
}

func decodeStringSlice(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return out
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getSanatoriumCapacity(ctx context.Context, q queryer, sanatoriumID string) (int, error) {
	const query = `SELECT total_places FROM deal.sanatoriums WHERE id = $1`
	var capacity int
	err := q.QueryRowContext(ctx, query, sanatoriumID).Scan(&capacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSanatoriumNotFound
		}
		return 0, err
	}
	return capacity, nil
}

func isSanatoriumAvailable(ctx context.Context, q queryer, sanatoriumID string, checkIn, checkOut time.Time, requestedGuests int, excludeBookingID *string) (bool, error) {
	query := `
SELECT COALESCE(SUM(b.guests), 0)
FROM deal.bookings b
WHERE b.sanatorium_id = $1
  AND b.status = 'confirmed'
  AND daterange(b.check_in, b.check_out, '[)') && daterange($2::date, $3::date, '[)')`
	args := []any{sanatoriumID, checkIn, checkOut}
	if excludeBookingID != nil && *excludeBookingID != "" {
		query += ` AND b.id <> $4`
		args = append(args, *excludeBookingID)
	}

	totalPlaces, err := getSanatoriumCapacity(ctx, q, sanatoriumID)
	if err != nil {
		return false, err
	}
	if requestedGuests <= 0 {
		requestedGuests = 1
	}

	var overlappingGuests int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&overlappingGuests); err != nil {
		return false, err
	}
	return hasCapacity(totalPlaces, overlappingGuests, requestedGuests), nil
}

func hasCapacity(totalPlaces, overlappingGuests, requestedGuests int) bool {
	if totalPlaces <= 0 || requestedGuests <= 0 {
		return false
	}
	return overlappingGuests+requestedGuests <= totalPlaces
}
