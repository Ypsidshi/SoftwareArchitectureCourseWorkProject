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

type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error %d: %s", e.StatusCode, e.Message)
}

func parseAuthError(status int, body []byte) error {
	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		msg := payload.Error
		if msg == "" {
			msg = payload.Message
		}
		if msg != "" {
			return &AuthError{StatusCode: status, Message: msg}
		}
	}
	return &AuthError{StatusCode: status, Message: string(body)}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	User        map[string]any `json:"user"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type RegisterResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
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
		return LoginResponse{}, parseAuthError(resp.StatusCode, payload)
	}

	var out LoginResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return LoginResponse{}, fmt.Errorf("decode login response: %w", err)
	}
	return out, nil
}

func (c *Client) Register(ctx context.Context, traceID string, req RegisterRequest) (RegisterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return RegisterResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/users/register", bytes.NewBuffer(body))
	if err != nil {
		return RegisterResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if traceID != "" {
		httpReq.Header.Set("X-Trace-Id", traceID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return RegisterResponse{}, err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return RegisterResponse{}, parseAuthError(resp.StatusCode, payload)
	}

	var out RegisterResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return RegisterResponse{}, fmt.Errorf("decode register response: %w", err)
	}
	return out, nil
}
