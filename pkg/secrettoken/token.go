package secrettoken

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/mr-tron/base58"
)

// Size in bytes
const Size = 32

// Token represents a secret token
type Token [Size]byte

const (
	// The index in a Token where flags are stored.
	// Not stored at beginning or end to prevent all secrets from starting or ending
	// with the same character (which would make them less easy to distinguish).
	flagsIndex = 16
)

// Nil is the empty token
var Nil = Token{}

// New creates a random Token with the flags byte embedded
func New(flags byte) Token {
	// generate random bytes
	dest := Token{}
	if _, err := rand.Read(dest[:]); err != nil {
		panic(err.Error())
	}

	// set flags
	dest[flagsIndex] = flags

	// done
	return Token(dest)
}

// FromString creates a Token from a string representation
func FromString(tokenString string) (t Token, err error) {
	bytes, err := base58.Decode(tokenString)
	if err != nil {
		return Nil, err
	}
	if len(bytes) != Size {
		return Nil, fmt.Errorf("Token is invalid (expected %d bytes)", Size)
	}
	copy(t[:], bytes)
	return t, nil
}

// FromStringOrNil creates a Token from a string representation
func FromStringOrNil(tokenString string) Token {
	t, err := FromString(tokenString)
	if err != nil {
		return Nil
	}
	return t
}

// String returns the string representation
func (t Token) String() string {
	return base58.Encode(t[:])
}

// Hashed returns a safe hash representation of the token
func (t Token) Hashed() string {
	// use sha256 digest
	hashed := sha256.Sum256(t[:])

	// encode hashed bytes as base62
	encoded := base58.Encode(hashed[:])

	// done
	return encoded
}

// Prefix returns a prefix to use to distinguish tokens
func (t Token) Prefix() string {
	s := t.String()
	if len(s) > 4 {
		return s[0:4]
	}
	return ""
}

// Flags returns my flags byte
func (t Token) Flags() byte {
	return t[flagsIndex]
}
