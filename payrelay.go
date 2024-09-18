package payrelay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var base = "https://api.payrelay.dev/2024-06-19"

type LNURLClient struct {
	client *Client
}

type Client struct {
	cfg *Config
}

var ErrNotFound = errors.New("not found")

func (c *Client) Fetch(ctx context.Context, method string, path string, body any, value any) error {
	// Combine the path with the base.
	j, err := url.JoinPath(base, path)
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)

	// Handle request bodys.
	if body != nil {
		// Encode the body.
		err = json.NewEncoder(b).Encode(body)
		if err != nil {
			return err
		}
	}

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, method, j, b)
	if err != nil {
		return err
	}

	// Set the authorization header.
	req.Header.Set("Authorization", "Bearer "+c.cfg.Secret)

	// Accept and return JSON.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the request.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Prevent a memory leak.
	defer res.Body.Close()

	// Handle error messages.
	if res.StatusCode != http.StatusOK {
		// Handle the specical case of not found, where no error message is needed.
		if res.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}

		// HACK: define the error message inline.
		var v struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		// Decode the error message.
		err := json.NewDecoder(res.Body).Decode(&v)
		if err != nil {
			return err
		}

		// Wrap the error message.
		return errors.New(v.Error.Message)
	}

	// Decode the response body.
	err = json.NewDecoder(res.Body).Decode(&value)
	if err != nil {
		return err
	}

	return nil
}

type InvoiceConfig struct {
	Amount int `json:"amount"`
}

type Invoice struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Amount int    `json:"amount"`
	PayReq string `json:"payreq"`
}

func (c *Client) NewInvoice(ctx context.Context, cfg *InvoiceConfig) (*Invoice, error) {
	var inv Invoice

	err := c.Fetch(ctx, "POST", "/invoice/create", cfg, &inv)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func (c *Client) QueryInvoice(ctx context.Context, id string) (*Invoice, error) {
	var inv Invoice

	p, err := url.JoinPath("invoice", id)
	if err != nil {
		return nil, err
	}

	err = c.Fetch(ctx, "GET", p, nil, &inv)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

type LNURLWConfig struct {
	Amount      int    `json:"amount"`
	WebhookURL  string `json:"webhookURL"`
	Description string `json:"description"`
}

type LNURLWState int

const (
	LNURLWStateReady LNURLWState = iota
	LNURLWStateScanned
	LNURLWStateCallback
)

type LNURLW struct {
	LNURL string
	ID    string
	State LNURLWState
}

func (c *Client) NewLNURLW(ctx context.Context, cfg *LNURLWConfig) (*LNURLW, error) {
	var w struct {
		LNURL string `json:"lnurl"`
		ID    string `json:"id"`
	}

	err := c.Fetch(ctx, "POST", "/lnurl/withdrawal/create", cfg, &w)
	if err != nil {
		return nil, err
	}

	return &LNURLW{
		LNURL: w.LNURL,
		ID:    w.ID,
		State: LNURLWStateReady,
	}, nil
}

func (c *Client) QueryLNURLW(ctx context.Context, id string) (*LNURLW, error) {
	var w struct {
		LNURL string `json:"lnurl"`
		ID    string `json:"id"`
		State string `json:"state"`
	}

	p, err := url.JoinPath("lnurl", "withdrawal", id)
	if err != nil {
		return nil, err
	}

	err = c.Fetch(ctx, "GET", p, nil, &w)
	if err != nil {
		return nil, err
	}

	var s LNURLWState

	switch w.State {
	case "READY":
		s = LNURLWStateReady

	case "SCANNED":
		s = LNURLWStateScanned

	case "CALLBACK":
		s = LNURLWStateCallback

	default:
		return nil, fmt.Errorf("unknown state: %s", w.State)
	}

	return &LNURLW{
		LNURL: w.LNURL,
		ID:    w.ID,
		State: s,
	}, nil
}

func (c *Client) DeleteLNURLW(ctx context.Context, id string) error {
	var res any

	p, err := url.JoinPath("lnurl", "withdrawal", id, "delete")
	if err != nil {
		return err
	}

	err = c.Fetch(ctx, "POST", p, nil, &res)
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	Secret string
}

func New(cfg *Config) *Client {
	return &Client{
		cfg: cfg,
	}
}
