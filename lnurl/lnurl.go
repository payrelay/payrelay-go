package lnurl

import (
	"context"
	"net/url"
)

type Provider interface {
	Fetch(ctx context.Context, method string, path string, body any, value any) error
}

type Client struct {
	provdier Provider
}

type WithdrawalConfig struct {
	Amount      int    `json:"amount"`
	Description string `json:"description"`
}

type Withdrawal struct {
	LNURL string `json:"lnurl"`
	ID    string `json:"id"`
	State string `json:"state"`
}

func (c *Client) NewWithdrawal(ctx context.Context, cfg *WithdrawalConfig) (*Withdrawal, error) {
	var w Withdrawal

	err := c.provdier.Fetch(ctx, "POST", "/lnurl/withdrawal/create", cfg, &w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (c *Client) QueryWithdrawal(ctx context.Context, id string) (*Withdrawal, error) {
	var w Withdrawal

	p, err := url.JoinPath("lnurl", "withdrawal", id)
	if err != nil {
		return nil, err
	}

	err = c.provdier.Fetch(ctx, "GET", p, nil, &w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (c *Client) DeleteWithdrawal(ctx context.Context, id string) error {
	var res any

	p, err := url.JoinPath("lnurl", "withdrawal", id, "delete")
	if err != nil {
		return err
	}

	err = c.provdier.Fetch(ctx, "POST", p, nil, &res)
	if err != nil {
		return err
	}

	return nil
}

func New(p Provider) *Client {
	return &Client{
		provdier: p,
	}
}
