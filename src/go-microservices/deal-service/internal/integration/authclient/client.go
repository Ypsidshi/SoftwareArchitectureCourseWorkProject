package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	User        map[string]any `json:"user"`
}

type Client struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Login(ctx context.Context, traceID string, req LoginRequest) (LoginResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return LoginResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return LoginResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if traceID != "" {
		httpReq.Header.Set("X-Trace-Id", traceID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return LoginResponse{}, fmt.Errorf("auth service returned %d: %s", resp.StatusCode, string(payload))
	}

	var out LoginResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return LoginResponse{}, fmt.Errorf("decode login response: %w", err)
	}
	return out, nil
}
