package hash

import (
	"errors"
	"log/slog"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type TokenHasher interface {
	Hash(token int64) (string, error)
}

type UrlTokenHash struct {
	logger *slog.Logger
}

func NewUrlTokenHash(logger *slog.Logger) *UrlTokenHash {
	return &UrlTokenHash{logger}
}

func (h UrlTokenHash) Hash(n int64) (string, error) {
	if n < 0 {
		return "", errors.New("invalid token provided")
	}

	base := int64(len(base62Chars)) // 62
	result := ""

	for n >= 0 {
		remainder := n % base
		result = string(base62Chars[remainder]) + result
		n = n/base - 1
		if n < 0 {
			break
		}
	}

	return result, nil
}

type FakeTokenHasher struct {
	HashFn func(n int64) (string, error)
}

func (f *FakeTokenHasher) Hash(n int64) (string, error) {
	return f.HashFn(n)
}
