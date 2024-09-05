package omada

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	defaultTimeout  = 30 * time.Second
	errTokenExpired = errors.New("token expired/unauthorized")
)

// Client connects to an Omada Controller.
type Client struct {
	logger     *zap.Logger
	http       *http.Client
	configPath string
	username   string
	password   string
	baseURL    string // protected
	token      string // protected
	mu         sync.Mutex
}

// NewClient returns a client that talks an Omada Controller.
func NewClient(logger *zap.Logger, config *Config) (*Client, error) {
	transport := &http.Transport{}
	if !config.Secure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c := Client{
		logger: logger,
		http: &http.Client{
			Transport: transport,
			Jar:       jar,
			Timeout:   defaultTimeout,
		},
		configPath: config.Path,
		baseURL:    strings.TrimSuffix(config.Path, "/"),
		username:   config.Username,
		password:   config.Password,
	}

	if err = c.authenticate(); err != nil {
		return nil, err
	}

	return &c, nil
}

// Token returns the auth token for the controller.  Some URLs may need it.
func (c *Client) Token() string {
	c.mu.Lock()
	t := c.token
	c.mu.Unlock()
	return t
}

func (c *Client) SetToken(token string) {
	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
}

// BaseURL returns the path to the Omada controller.  When authenticated,
// this will include the Controller ID as of Omada 5.x.
func (c *Client) BaseURL() string {
	c.mu.Lock()
	u := c.baseURL
	c.mu.Unlock()
	return u
}

// SetBaseURL updates the path to the Omada controller.
func (c *Client) SetBaseURL(path string) error {
	newPath, err := url.JoinPath(c.configPath, path)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.baseURL = strings.TrimSuffix(newPath, "/")
	c.mu.Unlock()

	return nil
}

func (c *Client) postJSON(url string, body io.Reader, target interface{}) error {
	req, err := http.NewRequest("POST", c.BaseURL()+url, body)
	if err != nil {
		return err
	}
	return c.doJSON(req, target)
}

func (c *Client) getJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", c.BaseURL()+url, nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, target)
}

func (c *Client) doJSON(req *http.Request, target interface{}) error {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Csrf-Token", c.Token())

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = io.ReadAll(res.Body)
		res.Body.Close()
	}()

	if res.StatusCode == http.StatusUnauthorized {
		return errTokenExpired
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%q returned %q", req.URL, res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return err
	}

	return nil
}

// retryOnce is a helper to call a function (typically make a HTTP request),
// and if it returns errTokenExpired, reauthenticate and try one more time.
func (c *Client) retryOnce(try func() error) error {
	retried := false

retry:
	err := try()
	if errors.Is(err, errTokenExpired) && !retried {
		retried = true
		if err = c.authenticate(); err != nil {
			return err
		}
		goto retry
	}
	return err
}

// authenticate updates c.token or returns an error.
func (c *Client) authenticate() error {

	type infoResult struct {
		ErrorCode int64  `json:"errorCode"`
		Msg       string `json:"msg"`
		Result    struct {
			ControllerID string `json:"omadacId"`
		} `json:"result"`
	}

	// Don't use old tokens during reauthentication.
	c.SetToken("")

	// Remove any known Controller ID.
	var err error
	err = c.SetBaseURL("/")
	if err != nil {
		return err
	}

	var ir infoResult
	err = c.getJSON("/api/info", &ir)
	if err != nil {
		return err
	}

	if ir.Result.ControllerID == "" {
		return fmt.Errorf("missing controller ID: %v: %q", ir.ErrorCode, ir.Msg)
	}

	err = c.SetBaseURL(ir.Result.ControllerID)
	if err != nil {
		return err
	}

	type authResult struct {
		ErrorCode int64  `json:"errorCode"`
		Msg       string `json:"msg"`
		Result    struct {
			RoleType int64  `json:"roleType"`
			Token    string `json:"token"`
		} `json:"result"`
	}

	kv := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	data, err := json.Marshal(kv)
	if err != nil {
		return err
	}

	var ar authResult
	err = c.postJSON("/api/v2/login", bytes.NewReader(data), &ar)
	if err != nil {
		return err
	}

	if ar.Result.Token == "" {
		return fmt.Errorf("auth failed: %v: %q", ar.ErrorCode, ar.Msg)
	}

	c.SetToken(ar.Result.Token)

	return nil
}
