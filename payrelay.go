package payrelay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/payrelay/payrelay-go/lnurl"
)

var base = "https://api.payrelay.dev/2024-06-19"

type LNURLClient struct {
	client *Client
}

type Client struct {
	cfg *Config
}

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

func (c *Client) LNURL() *lnurl.Client {
	return lnurl.New(c)
}

type Config struct {
	Secret string
}

func New(cfg *Config) *Client {
	return &Client{
		cfg: cfg,
	}
}
