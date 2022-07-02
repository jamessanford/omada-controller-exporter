package omada

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
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
	logger   *zap.Logger
	http     *http.Client
	baseURL  string
	username string
	password string
	token    string // protected
	tokenMu  sync.Mutex
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
		baseURL:  strings.TrimSuffix(config.Path, "/"),
		username: config.Username,
		password: config.Password,
	}

	if err = c.authenticate(); err != nil {
		return nil, err
	}

	return &c, nil
}

// Token returns the auth token for the controller.  Some URLs may need it.
func (c *Client) Token() string {
	c.tokenMu.Lock()
	t := c.token
	c.tokenMu.Unlock()
	return t
}

func (c *Client) postJSON(url string, body io.Reader, target interface{}) error {
	req, err := http.NewRequest("POST", c.baseURL+url, body)
	if err != nil {
		return err
	}
	return c.doJSON(req, target)
}

func (c *Client) getJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+url, nil)
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

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = ioutil.ReadAll(res.Body)
		res.Body.Close()
	}()

	if res.StatusCode == http.StatusUnauthorized {
		return errTokenExpired
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("returned %q", res.Status)
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

	c.tokenMu.Lock()
	c.token = ar.Result.Token
	c.tokenMu.Unlock()
	return nil
}
