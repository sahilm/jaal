package jaal

import (
	"crypto/sha256"
	"encoding/hex"
)

func CalculateSHA256(b []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ShortSHA256(b []byte) (string, error) {
	sha, err := CalculateSHA256(b)
	if err != nil {
		return "", err
	}
	return sha[:7], nil
}
