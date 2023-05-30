package sha512

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
)

type Gateway interface {
	SignHMACSHA512(cxt context.Context, text, key string) (string, error)
}

// compile time check that gateway implements Gateway interface
var _ Gateway = (*gateway)(nil)

// New is a constructor for the Gateway Interface that is provided to the fx
func New() (Gateway, error) {
	return &gateway{}, nil
}

type gateway struct{}

// SignHMACSHA512512 generates SHA512 hash for the text signed with key provided
func (g gateway) SignHMACSHA512(cxt context.Context, text, key string) (string, error) {
	mac := hmac.New(sha512.New, []byte(key))
	mac.Write([]byte(text))
	hmac := mac.Sum(nil)
	return fmt.Sprintf("%x", hmac), nil
}
