package httptransport

// swagger:model
type loginResponse struct {
	AccessToken string         `json:"access_token"`
	User        map[string]any `json:"user"`
}

// swagger:model
type medicalProfilesResponse struct {
	Items []string `json:"items"`
}

// swagger:model
type statusUpdatedResponse struct {
	Status string `json:"status"`
}

// swagger:model
type deletedResponse struct {
	Deleted bool `json:"deleted"`
}
