package domain

import "time"

type Sanatorium struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	City            string    `json:"city"`
	Address         string    `json:"address"`
	DistanceToSeaKM float64   `json:"distance_to_sea_km"`
	Amenities       []string  `json:"amenities"`
	ImageURLs       []string  `json:"image_urls"`
	PricePerNight   float64   `json:"price_per_night"`
	TotalPlaces     int       `json:"total_places"`
	MedicalProfiles []string  `json:"medical_profiles"`
	Latitude        *float64  `json:"latitude,omitempty"`
	Longitude       *float64  `json:"longitude,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
