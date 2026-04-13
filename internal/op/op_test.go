package op

import (
	"context"
	"testing"
)

// fakeClient is a test double for Client.
type fakeClient struct {
	accounts []Account
	err      error
	itemErr  error
}

func (f *fakeClient) AccountList(_ context.Context) ([]Account, error) {
	return f.accounts, f.err
}

func (f *fakeClient) ItemGet(_ context.Context, _ string) error {
	return f.itemErr
}

func TestFakeClientAccountList(t *testing.T) {
	want := []Account{{URL: "my.1password.com", Shorthand: "personal"}}
	c := &fakeClient{accounts: want}
	got, err := c.AccountList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].URL != want[0].URL {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFakeClientBinaryMissing(t *testing.T) {
	c := &fakeClient{err: ErrBinaryMissing}
	_, err := c.AccountList(context.Background())
	if err != ErrBinaryMissing {
		t.Errorf("got %v, want ErrBinaryMissing", err)
	}
}

func TestParseAccountJSON(t *testing.T) {
	data := []byte(`[{"url":"my.1password.com","shorthand":"personal"}]`)
	accounts, err := parseAccountJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("want 1 account, got %d", len(accounts))
	}
	if accounts[0].URL != "my.1password.com" {
		t.Errorf("got url %q", accounts[0].URL)
	}
}

func TestParseAccountJSONEmpty(t *testing.T) {
	accounts, err := parseAccountList([]byte("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("want 0 accounts, got %d", len(accounts))
	}
}
