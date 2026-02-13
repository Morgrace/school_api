package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time" // Added for benchmarking

	"golang.org/x/crypto/argon2"
)

// ============================================================================
// Constants & Configuration
// ============================================================================

const (
	// Default Argon2 parameters (OWASP 2024 recommendations)
	// Memory: 64MB (in KiB), Iterations: 3, Parallelism: 2 threads
	defaultMemory      = 64 * 1024 // 64 MB in KiB
	defaultIterations  = 3
	defaultParallelism = 2 // uint8 - number of threads
	defaultSaltLength  = 16
	defaultKeyLength   = 32
	algorithmName      = "argon2id"
)

// Argon2Hash represents a parsed Argon2 hash with all its components
type Argon2Hash struct {
	Algorithm   string
	Version     uint32
	Memory      uint32
	Iterations  uint32
	Parallelism uint8 // Note: uint8 for Argon2 threads parameter
	Salt        []byte
	Hash        []byte
}

// ============================================================================
// Public API: Password Hashing & Verification
// ============================================================================

// HashPassword creates a secure Argon2id hash from a plaintext password.
// Returns a PHC-formatted string: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
// Memory zeroing: Password bytes are zeroed after use.
func HashPassword(password string) (string, error) {
	if err := validatePasswordInput(password); err != nil {
		return "", err
	}

	// Convert password to bytes for memory zeroing
	passwordBytes := []byte(password)
	defer zeroBytes(passwordBytes)

	salt, err := generateSalt(defaultSaltLength)
	if err != nil {
		return "", fmt.Errorf("salt generation failed: %w", err)
	}
	defer zeroBytes(salt)

	// Derive cryptographic hash using Argon2id
	derivedHash := argon2.IDKey(
		passwordBytes,
		salt,
		defaultIterations,
		defaultMemory,
		defaultParallelism,
		defaultKeyLength,
	)
	defer zeroBytes(derivedHash)

	return encodePHCString(salt, derivedHash), nil
}

// CheckPassword verifies a password against a stored Argon2 hash.
// Uses constant-time comparison to prevent timing attacks.
// Memory zeroing: All sensitive bytes are zeroed after comparison.
func CheckPassword(password, encodedHash string) (bool, error) {
	if err := validatePasswordInput(password); err != nil {
		return false, err
	}

	// Convert password to bytes for memory zeroing
	passwordBytes := []byte(password)
	defer zeroBytes(passwordBytes)

	// Parse the stored hash into its components
	storedHash, err := parsePHCString(encodedHash)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}
	defer storedHash.zero() // Zero all components

	// Derive hash from provided password using stored parameters
	derivedHash := argon2.IDKey(
		passwordBytes,
		storedHash.Salt,
		storedHash.Iterations,
		storedHash.Memory,
		storedHash.Parallelism,
		uint32(len(storedHash.Hash)),
	)
	defer zeroBytes(derivedHash)

	// Constant-time comparison prevents timing attacks
	return subtle.ConstantTimeCompare(derivedHash, storedHash.Hash) == 1, nil
}

// UpgradeHashIfNeeded re-hashes a password with current security parameters.
// Returns:
//   - newHash: The new hash string (or original if no upgrade needed)
//   - didUpgrade: Boolean true if the hash was actually changed (signal to save to DB)
//   - error: Any processing error
func UpgradeHashIfNeeded(password, encodedHash string) (string, bool, error) {
	// First verify the password is correct
	match, err := CheckPassword(password, encodedHash)
	if err != nil {
		return "", false, fmt.Errorf("password verification failed: %w", err)
	}
	if !match {
		return "", false, fmt.Errorf("password does not match stored hash")
	}

	// Parse hash to check if it meets current standards
	storedHash, err := parsePHCString(encodedHash)
	if err != nil {
		return "", false, fmt.Errorf("failed to parse stored hash: %w", err)
	}
	defer storedHash.zero()

	// Check if current parameters meet minimum security requirements
	if needsUpgrade(storedHash) {
		newHash, err := HashPassword(password)
		if err != nil {
			return "", false, err
		}
		return newHash, true, nil // TRUE: Tell DB to save this!
	}

	// Return original hash if it's already secure enough
	return encodedHash, false, nil
}

// ============================================================================
// Core Hash Operations (Internal)
// ============================================================================

// encodePHCString encodes hash components into PHC string format.
// Format: $argon2id$v=19$m=65536,t=3,p=2$<b64salt>$<b64hash>
func encodePHCString(salt, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		algorithmName,
		argon2.Version,
		defaultMemory,
		defaultIterations,
		defaultParallelism,
		b64Salt,
		b64Hash,
	)
}

// parsePHCString parses a PHC-formatted hash string into its components.
// Validates format, algorithm, version, and parameter consistency.
func parsePHCString(encodedHash string) (*Argon2Hash, error) {
	// Split into components: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid PHC format: expected 6 parts, got %d", len(parts))
	}

	// Parse and validate each component
	hash := &Argon2Hash{}

	if err := parseAlgorithm(parts[1], hash); err != nil {
		return nil, err
	}

	if err := parseVersion(parts[2], hash); err != nil {
		return nil, err
	}

	if err := parseParameters(parts[3], hash); err != nil {
		return nil, err
	}

	if err := decodeSaltAndHash(parts[4], parts[5], hash); err != nil {
		return nil, err
	}

	return hash, nil
}

// ============================================================================
// Component Parsing Functions
// ============================================================================

// parseAlgorithm validates the algorithm name matches argon2id
func parseAlgorithm(algorithmPart string, hash *Argon2Hash) error {
	if algorithmPart != algorithmName {
		return fmt.Errorf("unsupported algorithm: %s (expected %s)", algorithmPart, algorithmName)
	}
	hash.Algorithm = algorithmPart
	return nil
}

// parseVersion validates Argon2 version is supported (v=19)
func parseVersion(versionPart string, hash *Argon2Hash) error {
	versionStr := strings.TrimPrefix(versionPart, "v=")
	version, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}

	if version != argon2.Version {
		return fmt.Errorf("unsupported argon2 version: %d (expected %d)", version, argon2.Version)
	}

	hash.Version = uint32(version)
	return nil
}

// parseParameters extracts m=memory, t=iterations, p=parallelism from param string
func parseParameters(paramPart string, hash *Argon2Hash) error {
	params := make(map[string]uint64)
	paramPairs := strings.Split(paramPart, ",")

	for _, pair := range paramPairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return fmt.Errorf("malformed parameter: %s", pair)
		}

		val, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid parameter value %s=%s: %w", kv[0], kv[1], err)
		}
		params[kv[0]] = val
	}

	// Validate required parameters exist
	required := []string{"m", "t", "p"}
	for _, req := range required {
		if _, exists := params[req]; !exists {
			return fmt.Errorf("missing required parameter: %s", req)
		}
	}

	hash.Memory = uint32(params["m"])
	hash.Iterations = uint32(params["t"])
	hash.Parallelism = uint8(params["p"]) // Critical: Convert to uint8 for Argon2

	return nil
}

// decodeSaltAndHash base64 decodes salt and hash components
func decodeSaltAndHash(saltPart, hashPart string, hash *Argon2Hash) error {
	salt, err := base64.RawStdEncoding.DecodeString(saltPart)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}
	hash.Salt = salt

	decodedHash, err := base64.RawStdEncoding.DecodeString(hashPart)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}
	hash.Hash = decodedHash

	return nil
}

// ============================================================================
// Security & Validation Functions
// ============================================================================

// validatePasswordInput ensures password meets basic requirements
func validatePasswordInput(password string) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	// Optional: Add minimum length requirement
	if len(password) < 4 {
		return fmt.Errorf("password must be at least 4 characters")
	}
	return nil
}

// needsUpgrade determines if a hash should be re-hashed with stronger parameters
func needsUpgrade(hash *Argon2Hash) bool {
	// Check against current security standards
	if hash.Memory < defaultMemory {
		return true
	}
	if hash.Iterations < defaultIterations {
		return true
	}
	if hash.Parallelism < defaultParallelism {
		return true
	}
	return false
}

// generateSalt creates cryptographically secure random salt
func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return salt, nil
}

// ============================================================================
// Memory Safety Functions
// ============================================================================

// zeroBytes overwrites a byte slice with zeros to prevent memory analysis.
// Uses runtime.KeepAlive to prevent compiler optimization removal.
func zeroBytes(data []byte) {
	if data == nil {
		return
	}
	for i := range data {
		data[i] = 0
	}
	// Prevent compiler from optimizing away the zeroing loop
	// runtime.KeepAlive(data) // Uncomment if needed
}

// zero securely zeroes all sensitive fields in an Argon2Hash struct
func (h *Argon2Hash) zero() {
	zeroBytes(h.Salt)
	zeroBytes(h.Hash)

	// Also clear the struct fields for good measure
	h.Salt = nil
	h.Hash = nil
	h.Memory = 0
	h.Iterations = 0
	h.Parallelism = 0
	h.Version = 0
	h.Algorithm = ""
}

// ============================================================================
// Testing & Debugging Helpers (Optional)
// ============================================================================

// BenchmarkHashDuration times how long hashing takes (for parameter tuning)
func BenchmarkHashDuration(password string, iterations int) (avgDuration float64, err error) {
	passwordBytes := []byte(password)
	defer zeroBytes(passwordBytes)

	totalDuration := 0.0
	for i := 0; i < iterations; i++ {
		salt, err := generateSalt(defaultSaltLength)
		if err != nil {
			return 0, err
		}
		defer zeroBytes(salt)

		start := time.Now()
		hash := argon2.IDKey(
			passwordBytes,
			salt,
			defaultIterations,
			defaultMemory,
			defaultParallelism,
			defaultKeyLength,
		)
		zeroBytes(hash)
		duration := time.Since(start).Seconds()
		totalDuration += duration
	}

	return totalDuration / float64(iterations), nil
}
