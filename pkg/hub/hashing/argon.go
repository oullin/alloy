package hashing

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

type keyFunc func(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte

// ArgonHasher hashes values using the argon2i algorithm.
type ArgonHasher struct {
	mu        sync.RWMutex
	memory    uint32
	time      uint32
	threads   uint8
	verify    bool
	algorithm string
	derive    keyFunc
}

// NewArgonHasher creates an ArgonHasher (argon2i). Recognised option keys:
// "memory" (uint32/int), "time" (uint32/int), "threads" (uint8/int), "verify" (bool).

// Info returns metadata about an argon2 hashed value.

// Make hashes a plaintext value.

// Check verifies a plaintext value against an argon2 hash.

// NeedsRehash reports whether the hashed value was created with different options.

// VerifyConfiguration checks whether the hash was produced with acceptable settings.

// SetMemory updates the memory cost.

// SetTime updates the time cost.

// SetThreads updates the parallelism.

// Memory returns the configured memory cost.

// Time returns the configured time cost.

// Threads returns the configured thread count.

type argonParams struct {
	algorithm string
	memory    uint32
	time      uint32
	threads   uint8
	salt      []byte
	hash      []byte
}

const (
	// argonDefaultMemory is the default memory cost in KiB. 19456 KiB = 19 MiB,
	// the minimum recommended by RFC 9106 (second recommended option) and OWASP
	// for argon2id. The previous 1 MiB default was ~19x below this floor and
	// left password hashes cheap to crack on GPU/ASIC hardware.
	argonDefaultMemory uint32 = 19 * 1024
	// argonDefaultTime is the default iteration (time) cost, matching the
	// OWASP/RFC 9106 recommendation of t=2 for the 19 MiB memory profile.
	argonDefaultTime uint32 = 2
	// argonDefaultThreads is the default parallelism (lanes). Two lanes meets
	// or exceeds the RFC 9106/OWASP p>=1 guidance.
	argonDefaultThreads uint8  = 2
	argonSaltLen               = 16
	argonHashLen        uint32 = 32
)

var _ Hasher = (*ArgonHasher)(nil)

func NewArgonHasher(opts ...map[string]any) *ArgonHasher {
	h := &ArgonHasher{
		memory:    argonDefaultMemory,
		time:      argonDefaultTime,
		threads:   argonDefaultThreads,
		verify:    false,
		algorithm: "argon2i",
		derive:    argon2.Key,
	}

	if len(opts) > 0 {
		applyArgonOpts(h, opts[0])
	}

	return h
}

func (h *ArgonHasher) Info(hashedValue string) (HashInfo, error) {
	params, err := parseArgonHash(hashedValue)

	if err != nil {
		return HashInfo{}, err
	}

	return HashInfo{
		Algorithm: params.algorithm,
		Options: map[string]any{
			"memory":  params.memory,
			"time":    params.time,
			"threads": params.threads,
		},
	}, nil
}

func (h *ArgonHasher) Make(value string, options ...map[string]any) (string, error) {
	memory, time, threads := h.params(options...)

	salt := make([]byte, argonSaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("hashing: %w", err)
	}

	hash := h.derive([]byte(value), salt, time, memory, threads, argonHashLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		h.algorithm, argon2.Version, memory, time, threads, b64Salt, b64Hash)

	return encoded, nil
}

func (h *ArgonHasher) Check(value string, hashedValue string, options ...map[string]any) (bool, error) {
	if len(hashedValue) == 0 {
		return false, nil
	}

	if h.verify && !h.isUsingCorrectAlgorithm(hashedValue) {
		return false, ErrAlgorithmMismatch
	}

	params, err := parseArgonHash(hashedValue)

	if err != nil {
		return false, nil
	}

	hash := h.derive([]byte(value), params.salt, params.time, params.memory, params.threads, uint32(len(params.hash)))

	if subtle.ConstantTimeCompare(hash, params.hash) != 1 {
		return false, nil
	}

	return true, nil
}

func (h *ArgonHasher) NeedsRehash(hashedValue string, options ...map[string]any) (bool, error) {
	params, err := parseArgonHash(hashedValue)

	if err != nil {
		return true, nil
	}

	memory, time, threads := h.params(options...)

	return params.memory != memory || params.time != time || params.threads != threads, nil
}

func (h *ArgonHasher) VerifyConfiguration(hashedValue string) bool {
	params, err := parseArgonHash(hashedValue)

	if err != nil {
		return false
	}

	memory, time, threads := h.params()

	return params.memory <= memory && params.time <= time && params.threads <= threads
}

func (h *ArgonHasher) SetMemory(memory uint32) {
	if memory == 0 {
		return
	}

	h.mu.Lock()

	defer h.mu.Unlock()

	h.memory = memory
}

func (h *ArgonHasher) SetTime(time uint32) {
	if time == 0 {
		return
	}

	h.mu.Lock()

	defer h.mu.Unlock()

	h.time = time
}

func (h *ArgonHasher) SetThreads(threads uint8) {
	if threads == 0 {
		return
	}

	h.mu.Lock()

	defer h.mu.Unlock()

	h.threads = threads
}

func (h *ArgonHasher) Memory() uint32 {
	h.mu.RLock()

	defer h.mu.RUnlock()

	return h.memory
}

func (h *ArgonHasher) Time() uint32 {
	h.mu.RLock()

	defer h.mu.RUnlock()

	return h.time
}

func (h *ArgonHasher) Threads() uint8 {
	h.mu.RLock()

	defer h.mu.RUnlock()

	return h.threads
}

func (h *ArgonHasher) isUsingCorrectAlgorithm(hashedValue string) bool {
	return strings.HasPrefix(hashedValue, "$"+h.algorithm+"$")
}

func (h *ArgonHasher) params(options ...map[string]any) (uint32, uint32, uint8) {
	h.mu.RLock()
	memory, time, threads := h.memory, h.time, h.threads
	h.mu.RUnlock()

	if len(options) > 0 {
		if v, ok := toUint32(options[0]["memory"]); ok {
			memory = v
		}

		if v, ok := toUint32(options[0]["time"]); ok {
			time = v
		}

		if v, ok := toUint8(options[0]["threads"]); ok {
			threads = v
		}
	}

	return memory, time, threads
}

func parseArgonHash(encoded string) (argonParams, error) {
	// Expected format: $argon2i$v=19$m=M,t=T,p=P$salt$hash
	parts := strings.Split(encoded, "$")

	if len(parts) != 6 {
		return argonParams{}, ErrInvalidHash
	}

	algo := parts[1]

	if algo != "argon2i" && algo != "argon2id" {
		return argonParams{}, ErrInvalidHash
	}

	var memory, time uint32

	var threads uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)

	if err != nil || memory == 0 || time == 0 || threads == 0 {
		return argonParams{}, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])

	if err != nil || len(salt) == 0 {
		return argonParams{}, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])

	if err != nil || len(hash) == 0 {
		return argonParams{}, ErrInvalidHash
	}

	return argonParams{
		algorithm: algo,
		memory:    memory,
		time:      time,
		threads:   threads,
		salt:      salt,
		hash:      hash,
	}, nil
}

func applyArgonOpts(h *ArgonHasher, opts map[string]any) {
	if v, ok := toUint32(opts["memory"]); ok {
		h.memory = v
	}

	if v, ok := toUint32(opts["time"]); ok {
		h.time = v
	}

	if v, ok := toUint8(opts["threads"]); ok {
		h.threads = v
	}

	if v, ok := opts["verify"].(bool); ok {
		h.verify = v
	}
}

func toUint32(v any) (uint32, bool) {
	switch val := v.(type) {
	case uint32:
		return val, val > 0
	case int:
		if val <= 0 {
			return 0, false
		}

		return uint32(val), true
	case int64:
		if val <= 0 {
			return 0, false
		}

		return uint32(val), true
	default:
		return 0, false
	}
}

func toUint8(v any) (uint8, bool) {
	switch val := v.(type) {
	case uint8:
		return val, val > 0
	case int:
		if val <= 0 {
			return 0, false
		}

		return uint8(val), true
	case int64:
		if val <= 0 {
			return 0, false
		}

		return uint8(val), true
	default:
		return 0, false
	}
}
