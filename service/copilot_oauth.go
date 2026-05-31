package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const (
	// GitHub's OAuth app used by VS Code for GitHub/Copilot device authorization.
	copilotDeviceClientID    = "Iv1.b507a08c87ecfe98"
	copilotDeviceCodeURL     = "https://github.com/login/device/code"
	copilotDeviceTokenURL    = "https://github.com/login/oauth/access_token"
	copilotDeviceScope       = "read:user"
	copilotDeviceHTTPTimeout = 20 * time.Second
)

type CopilotDeviceCodeResult struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
}

type CopilotDeviceTokenResult struct {
	AccessToken string
	TokenType   string
	Scope       string
}

func StartCopilotDeviceFlow(ctx context.Context, proxyURL string) (*CopilotDeviceCodeResult, error) {
	client, err := getCopilotDeviceHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("client_id", copilotDeviceClientID)
	form.Set("scope", copilotDeviceScope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, copilotDeviceCodeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
		Error                   string `json:"error"`
		ErrorDescription        string `json:"error_description"`
	}
	if err := common.DecodeJson(resp.Body, &payload); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || strings.TrimSpace(payload.Error) != "" {
		return nil, fmt.Errorf("copilot device flow start failed: status=%d error=%s description=%s", resp.StatusCode, payload.Error, payload.ErrorDescription)
	}
	if strings.TrimSpace(payload.DeviceCode) == "" || strings.TrimSpace(payload.UserCode) == "" || strings.TrimSpace(payload.VerificationURI) == "" {
		return nil, errors.New("copilot device flow response missing required fields")
	}
	if payload.Interval <= 0 {
		payload.Interval = 5
	}
	return &CopilotDeviceCodeResult{
		DeviceCode:              strings.TrimSpace(payload.DeviceCode),
		UserCode:                strings.TrimSpace(payload.UserCode),
		VerificationURI:         strings.TrimSpace(payload.VerificationURI),
		VerificationURIComplete: strings.TrimSpace(payload.VerificationURIComplete),
		ExpiresIn:               payload.ExpiresIn,
		Interval:                payload.Interval,
	}, nil
}

func CompleteCopilotDeviceFlow(ctx context.Context, deviceCode string, proxyURL string) (*CopilotDeviceTokenResult, error) {
	client, err := getCopilotDeviceHTTPClient(proxyURL)
	if err != nil {
		return nil, err
	}

	code := strings.TrimSpace(deviceCode)
	if code == "" {
		return nil, errors.New("empty device_code")
	}

	form := url.Values{}
	form.Set("client_id", copilotDeviceClientID)
	form.Set("device_code", code)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, copilotDeviceTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := common.DecodeJson(resp.Body, &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.Error) != "" {
		return nil, fmt.Errorf("copilot device authorization pending or failed: %s %s", payload.Error, payload.ErrorDescription)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("copilot device token exchange failed: status=%d", resp.StatusCode)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return nil, errors.New("copilot device token exchange returned empty access_token")
	}
	return &CopilotDeviceTokenResult{
		AccessToken: strings.TrimSpace(payload.AccessToken),
		TokenType:   strings.TrimSpace(payload.TokenType),
		Scope:       strings.TrimSpace(payload.Scope),
	}, nil
}

func getCopilotDeviceHTTPClient(proxyURL string) (*http.Client, error) {
	baseClient, err := GetHttpClientWithProxy(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if baseClient == nil {
		return &http.Client{Timeout: copilotDeviceHTTPTimeout}, nil
	}
	clientCopy := *baseClient
	clientCopy.Timeout = copilotDeviceHTTPTimeout
	return &clientCopy, nil
}
