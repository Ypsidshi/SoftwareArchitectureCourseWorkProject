package service

import (
	"context"
	"errors"
	"fmt"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/repository"
)

type SanatoriumInput struct {
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

type ListSanatoriumsAdminResult struct {
	Items      []domain.Sanatorium `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

type sanatoriumAdminRepo interface {
	bookingRepo
	CreateSanatorium(ctx context.Context, in repository.NewSanatorium) (domain.Sanatorium, error)
	UpdateSanatorium(ctx context.Context, in repository.UpdateSanatorium) (domain.Sanatorium, error)
	DeleteSanatorium(ctx context.Context, id string) error
	ListSanatoriumsAdmin(ctx context.Context, page, pageSize int) ([]domain.Sanatorium, int, error)
	ListMedicalProfileNames(ctx context.Context) ([]string, error)
}

func (s *BookingService) ListSanatoriumsAdmin(ctx context.Context, page, pageSize int) (ListSanatoriumsAdminResult, error) {
	page, pageSize = normalizePagination(page, pageSize)
	repo, ok := s.repo.(sanatoriumAdminRepo)
	if !ok {
		return ListSanatoriumsAdminResult{}, fmt.Errorf("sanatorium admin repo not supported")
	}
	items, total, err := repo.ListSanatoriumsAdmin(ctx, page, pageSize)
	if err != nil {
		return ListSanatoriumsAdminResult{}, err
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return ListSanatoriumsAdminResult{
		Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
	}, nil
}

func (s *BookingService) CreateSanatorium(ctx context.Context, in SanatoriumInput) (domain.Sanatorium, error) {
	if in.Name == "" || in.City == "" {
		return domain.Sanatorium{}, fmt.Errorf("name and city are required")
	}
	if in.PricePerNight <= 0 || in.TotalPlaces <= 0 {
		return domain.Sanatorium{}, fmt.Errorf("price_per_night and total_places must be positive")
	}
	repo := s.repo.(sanatoriumAdminRepo)
	return repo.CreateSanatorium(ctx, repository.NewSanatorium{
		Name: in.Name, Description: in.Description, City: in.City, Address: in.Address,
		DistanceToSeaKM: in.DistanceToSeaKM, Amenities: in.Amenities, ImageURLs: in.ImageURLs,
		PricePerNight: in.PricePerNight, TotalPlaces: in.TotalPlaces,
		Latitude: in.Latitude, Longitude: in.Longitude, MedicalProfiles: in.MedicalProfiles,
	})
}

func (s *BookingService) UpdateSanatorium(ctx context.Context, in SanatoriumInput, id string) (domain.Sanatorium, error) {
	if id == "" {
		return domain.Sanatorium{}, fmt.Errorf("id is required")
	}
	if in.PricePerNight <= 0 || in.TotalPlaces <= 0 {
		return domain.Sanatorium{}, fmt.Errorf("price_per_night and total_places must be positive")
	}
	repo := s.repo.(sanatoriumAdminRepo)
	return repo.UpdateSanatorium(ctx, repository.UpdateSanatorium{
		ID: id, Name: in.Name, Description: in.Description, City: in.City, Address: in.Address,
		DistanceToSeaKM: in.DistanceToSeaKM, Amenities: in.Amenities, ImageURLs: in.ImageURLs,
		PricePerNight: in.PricePerNight, TotalPlaces: in.TotalPlaces,
		Latitude: in.Latitude, Longitude: in.Longitude, MedicalProfiles: in.MedicalProfiles,
	})
}

func (s *BookingService) ListMedicalProfileNames(ctx context.Context) ([]string, error) {
	repo, ok := s.repo.(sanatoriumAdminRepo)
	if !ok {
		return nil, fmt.Errorf("sanatorium admin repo not supported")
	}
	return repo.ListMedicalProfileNames(ctx)
}

func (s *BookingService) DeleteSanatorium(ctx context.Context, id string) error {
	return s.repo.(sanatoriumAdminRepo).DeleteSanatorium(ctx, id)
}

func IsSanatoriumHasBookings(err error) bool {
	return errors.Is(err, repository.ErrSanatoriumHasBookings)
}
