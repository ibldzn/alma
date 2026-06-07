package fincloud

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL    = "https://172.20.57.7/fincloud-taspen-web"
	DefaultUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:142.0) Gecko/20100101 Firefox/142.0"
	DefaultAPIBaseURL = "https://172.20.57.5:17000"
	loginPath         = "/admin/access/login"
	logoutPath        = "/admin/access/logout"
)

var errMissingCredentials = errors.New("missing Fincloud credentials")

// Config defines the parameters for constructing a Client.
// LocationID and RoleID are required by the upstream Fincloud login endpoint.
type Config struct {
	BaseURL    string
	BaseURLAPI string
	UserAgent  string
	APIKey     string
	HTTPClient *http.Client
}

// Credentials bundles the username/password required to login.
type Credentials struct {
	Username   string
	Password   string
	LocationID string
	RoleID     string
}

// Client wraps Fincloud specific HTTP interactions.
type Client struct {
	cfg Config
}

// Session captures the state returned by a successful login.
type Session struct {
	ID           string
	LocationID   string
	LocationName string
	RoleID       string
	RoleName     string
	IdleTimeout  time.Duration
	IdleWarning  time.Duration
	raw          loginResponse
}

// LoginResponse exposes the upstream payload; exported for debugging when needed.
type LoginResponse struct {
	loginResponse
}

type AuthorizationModel struct {
	Locations []AuthLabel `json:"locationid"`
	Roles     []AuthLabel `json:"roleid"`
}

type AuthLabel struct {
	ID          string `json:"id"`
	Description string `json:"descr"`
}

// NewClient constructs a Fincloud client with sane defaults.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	if cfg.BaseURLAPI == "" {
		cfg.BaseURLAPI = DefaultAPIBaseURL
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}

	defaultTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout:   2 * time.Minute,
			Transport: defaultTransport,
		}
	} else {
		httpClient := *cfg.HTTPClient
		if httpClient.Timeout == 0 {
			httpClient.Timeout = 2 * time.Minute
		}
		if httpClient.Transport == nil {
			httpClient.Transport = defaultTransport
		}
		cfg.HTTPClient = &httpClient
	}

	return &Client{cfg}, nil
}

// Login authenticates the provided credentials and returns a Session handle.
func (c *Client) Login(ctx context.Context, cred Credentials) (*Session, error) {
	if strings.TrimSpace(cred.Username) == "" || strings.TrimSpace(cred.Password) == "" {
		return nil, errMissingCredentials
	}

	form := url.Values{}
	form.Set("locationid", cred.LocationID)
	form.Set("roleid", cred.RoleID)
	form.Set("username", cred.Username)
	form.Set("pwd", cred.Password)

	req, err := c.newRequest(ctx, http.MethodPost, loginPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed: %s", resp.Status)
	}

	var payload loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if payload.Status != "ok" {
		if payload.Error != nil {
			return nil, fmt.Errorf("login error: %s", payload.Error.System)
		}
		return nil, errors.New("login failed with unknown error")
	}

	session := Session{
		ID:           payload.Data.Result.SessionID,
		LocationID:   payload.Data.Result.LocationID,
		LocationName: payload.Data.Result.LocationName,
		RoleID:       payload.Data.Result.RoleID,
		RoleName:     payload.Data.Result.RoleName,
		IdleTimeout:  time.Duration(payload.Data.Result.IdleTimeout) * time.Second,
		IdleWarning:  time.Duration(payload.Data.Result.IdleWarning) * time.Second,
		raw:          payload,
	}

	return &session, nil
}

func (c *Client) Logout(ctx context.Context, sessionID string) error {
	req, err := c.newRequest(ctx, http.MethodPost, logoutPath, nil)
	if err != nil {
		return err
	}
	AttachSession(req, sessionID)

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed: %s", resp.Status)
	}

	return nil
}

func (c *Client) GetAuthLabels(ctx context.Context) (*AuthorizationModel, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/admin/access/listvalues", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var intermediate struct {
		Data struct {
			Result AuthorizationModel `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intermediate); err != nil {
		return nil, err
	}

	if intermediate.Status != "ok" {
		return nil, fmt.Errorf("failed to fetch authorization labels")
	}

	return &intermediate.Data.Result, nil
}

// Do executes the provided request using the configured HTTP client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.cfg.HTTPClient.Do(req)
}

// NewRequest prepares a Fincloud HTTP request with common headers applied.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// newRequest builds an http.Request anchored to the configured BaseURL.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	return req, nil
}

// AttachSession sets the session header required by authenticated endpoints.
func AttachSession(req *http.Request, sessionID string) {
	req.Header.Set("sessionid", sessionID)
}

// Raw exposes the original login payload for troubleshooting purposes.
func (s Session) Raw() LoginResponse {
	return LoginResponse{loginResponse: s.raw}
}

// loginResponse mirrors the JSON payload returned by the login endpoint.
type loginResponse struct {
	Data struct {
		Result struct {
			IdleTimeout  int64  `json:"idletimeout"`
			IdleWarning  int64  `json:"idlewarning"`
			LocationID   string `json:"locationid"`
			LocationName string `json:"locationname"`
			RoleID       string `json:"roleid"`
			RoleName     string `json:"rolename"`
			SessionID    string `json:"sessionid"`
		} `json:"result"`
	} `json:"data"`
	Error *struct {
		System string `json:"system"`
	} `json:"error,omitempty"`
	Status string `json:"status"`
}

func (c *Client) DownloadReportFromMaintenance(ctx context.Context, file, path string) (string, error) {
	// /system/downloaderlaporan/download.php?file=cbrcustomer.csv&path=/app/report/cbr/20251231
	sessionId, ok := SessionIDFromContext(ctx)
	if !ok || sessionId == "" {
		return "", errors.New("missing session ID in context")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/system/downloaderlaporan/download.php", nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("file", file)
	q.Set("path", path)
	req.URL.RawQuery = q.Encode()

	AttachSession(req, sessionId)

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download report failed: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")) // remove BOM if present

	return string(data), nil
}

// DecodeJSON is a helper that unmarshals the response into v and surfaces HTTP status issues.
func DecodeJSON(resp *http.Response, v any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fincloud request failed: %s", resp.Status)
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// computeSignature computes the HMAC-SHA256 signature for the given payload and secret.
func (c *Client) computeSignature(payload []byte) string {
	hashed := hmac.New(sha256.New, []byte(c.cfg.APIKey))
	hashed.Write(payload)
	return hex.EncodeToString(hashed.Sum(nil))
}

// NewAPIRequest prepares a Fincloud API HTTP request with common headers applied.
func (c *Client) NewAPIRequest(ctx context.Context, method, path string, jsonBody []byte) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := c.cfg.BaseURLAPI + path

	var bodyReader io.Reader
	if jsonBody != nil {
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	if jsonBody != nil {
		if c.cfg.APIKey == "" {
			return nil, errMissingCredentials
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("signature", c.computeSignature(jsonBody))
	}

	return req, nil
}

func (c *Client) DownloadReport(ctx context.Context, name string, params ...any) (string, error) {
	sessionId, ok := SessionIDFromContext(ctx)
	if !ok || sessionId == "" {
		return "", errors.New("missing session ID in context")
	}

	p := make([]string, len(params))
	for i, v := range params {
		switch t := v.(type) {
		case string:
			p[i] = t
		case time.Time:
			p[i] = t.Format("2006-01-02")
		default:
			p[i] = fmt.Sprint(t)
		}
	}

	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("nm", name)
	q.Set("type", "csv")
	q.Set("p", string(b))

	path := "/system/laporanUmum/data/lap?" + q.Encode()
	path = strings.ReplaceAll(path, "+", "%20") // space encoding

	form := url.Values{}
	form.Set("sessionId", sessionId)

	req, err := c.newRequest(ctx, http.MethodGet, path, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	AttachSession(req, sessionId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download report failed: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")) // remove BOM if present

	return string(data), nil
}
