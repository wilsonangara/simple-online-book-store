package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func Test_NewClient(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		testValidSecret := "valid-secret"

		c, err := NewClient("valid-secret")
		if err != nil {
			t.Fatalf("NewClient(_), expected nil error, got = %v", err)
		}

		if c.secret != testValidSecret {
			t.Fatalf("NewClient(_) error, got = %s, want = %s", c.secret, testValidSecret)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		_, err := NewClient("")
		if err != ErrSecretIsRequired {
			t.Fatalf("NewStorage(_) error, got = %v, want = %v", err, ErrSecretIsRequired)
		}
	})
}

func Test_GenerateToken(t *testing.T) {
	t.Parallel()

	var (
		validSecret = uuid.New().String()
		validID     = int64(1)
	)

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			secret: validSecret,
		}

		_, err := c.GenerateToken(validID)
		if err != nil {
			t.Fatalf("GenerateToken(_) expected nil error, got = %v", err)
		}
	})

	t.Run("Failed", func(t *testing.T) {
		t.Parallel()

		wantErr := errIDIsRequired

		c := &Client{
			secret: validSecret,
		}

		_, err := c.GenerateToken(0)
		if !errors.Is(err, wantErr) {
			t.Fatalf("GenerateToken(_) error, got = %v, want = %v", err, wantErr)
		}
	})
}

func Test_ValidateToken(t *testing.T) {
	t.Parallel()

	var (
		validSecret = "valid-secret"
		validID     = int64(1)
	)

	c := &Client{
		secret: validSecret,
	}

	// generate a valid token.
	token, err := c.GenerateToken(validID)
	if err != nil {
		t.Fatalf("GenerateToken(_) unexpected error when generating token: %v", err)
	}

	// check acess token
	id, err := c.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken(_) expected nil error, got = %v", err)
	}

	if id != validID {
		t.Fatalf("ValidateToken(_) error, got = %v, want = %v", id, validID)
	}
}
