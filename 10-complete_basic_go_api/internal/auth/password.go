package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// OWASP minimum Argon2id parameters. Memory is KiB.
const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19 * 1024
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

type Argon2 struct{}

func (Argon2) Hash(password string) (string, error) {
	return HashPassword(password)
}

func (Argon2) Verify(password, encoded string) (bool, error) {
	return VerifyPassword(password, encoded)
}

func HashPassword(password string) (string, error) {
	if password == "" || len(password) > 256 {
		return "", errors.New("invalid password length")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	timeCost, memory, threads, salt, hash, err := parseHash(encoded)
	if err != nil {
		return false, err
	}
	if password == "" || len(password) > 256 {
		return false, nil
	}

	sum := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(hash)))
	if subtle.ConstantTimeCompare(sum, hash) == 1 {
		return true, nil
	}
	return false, nil
}

func parseHash(encoded string) (uint32, uint32, uint8, []byte, []byte, error) {
	parts := splitHash(encoded)
	if parts == nil {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash version")
	}
	if parts[2] != fmt.Sprintf("v=%d", version) {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash version")
	}

	var memory, timeCost, threads int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash parameters")
	}
	if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, timeCost, threads) {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash parameters")
	}
	if memory <= 0 || memory > 64*1024 || timeCost <= 0 || timeCost > 5 || threads <= 0 || threads > 4 {
		return 0, 0, 0, nil, nil, errors.New("unsupported password hash parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < saltLen {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash")
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) != int(argonKeyLen) {
		return 0, 0, 0, nil, nil, errors.New("invalid password hash")
	}

	return uint32(timeCost), uint32(memory), uint8(threads), salt, hash, nil
}

func splitHash(encoded string) []string {
	parts := make([]string, 0, 6)
	start := 0
	for i := 0; i <= len(encoded); i++ {
		if i == len(encoded) || encoded[i] == '$' {
			parts = append(parts, encoded[start:i])
			start = i + 1
		}
	}
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return nil
	}
	return parts
}
