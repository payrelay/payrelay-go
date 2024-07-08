package payrelay

import (
	"context"
	"testing"

	"github.com/payrelay/payrelay-go/lnurl"
)

func TestPayRelay(t *testing.T) {
	c := New(&Config{
		Secret: "00000000-0000-0000-0000-000000000000",
	})

	w, err := c.LNURL().NewWithdrawal(context.TODO(), &lnurl.WithdrawalConfig{
		Amount:      100,
		Description: "Hello World",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", w)
}
