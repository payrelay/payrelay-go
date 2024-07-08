package lnurl

import (
	"context"
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
}

func (c *Client) NewWithdrawal(ctx context.Context, cfg *WithdrawalConfig) (*Withdrawal, error) {
	var w Withdrawal

	err := c.provdier.Fetch(ctx, "POST", "/lnurl/withdrawal/create", cfg, &w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func New(p Provider) *Client {
	return &Client{
		provdier: p,
	}
}
