package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrFailDecodeHash        = errors.New("failed to decode hash")
	ErrFailDecodeSalt        = errors.New("failed to decode salt")
	ErrFailParseAttribute    = errors.New("failed to parse hash attribute")
	ErrIncompatibleAlgorithm = errors.New("incompatible hashing algorithm")
	ErrIncompatibleVersion   = errors.New("incompatible argon2 version")
	ErrInvalidHashFormat     = errors.New("invalid salted hash format")
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Salts and hashes a provided string using argon2id and the provided argon parameters
// Returns a salted argon2id hash
func HashSaltString(toHash string, params *Argon2Params) string {
	salt := make([]byte, params.SaltLength)
	rand.Read(salt) // nolint:errcheck
	// crypto/rand.Read "never" returns an error and if it fails, it crashes the program per the documentation

	hash := argon2.IDKey(
		[]byte(toHash),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		encodedSalt,
		encodedHash,
	)
}

// Parses a salted hash string and retrieves the argon params, salt, and hash
// Allows for the extraction of the parameters from the hash string
func DecodeSaltedHash(saltedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(saltedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHashFormat
	}
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrIncompatibleAlgorithm
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrFailParseAttribute
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	params := &Argon2Params{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return nil, nil, nil, ErrFailParseAttribute
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrFailDecodeSalt, err)
	}
	params.SaltLength = uint32(len(salt))
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrFailDecodeHash, err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// Securely compares a salted hash against a provided password and returns
// true if they match and false otherwise
func SecureVerifyHash(saltedHash string, password string) (bool, error) {
	params, salt, hash, err := DecodeSaltedHash(saltedHash)
	if err != nil {
		return false, err
	}

	compareHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, compareHash) == 1 {
		return true, nil
	}
	return false, nil
}
