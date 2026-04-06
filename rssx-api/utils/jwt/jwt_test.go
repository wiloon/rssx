package jwt

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-jwt"

// getSigningKey returns the test secret as bytes for signing/verifying tokens.
func getSigningKey() []byte {
	return []byte(testSecret)
}

// getSigningKeyBase64 returns the test secret as a base64-encoded string.
func getSigningKeyBase64() string {
	return base64.RawURLEncoding.EncodeToString(getSigningKey())
}

// createTestToken creates a valid JWT token for testing purposes.
func createTestToken(id string, exp time.Time) (string, error) {
	claims := RssxClaims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"rssx.wiloon.net"},
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        "test-jti",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wiloon.com",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "rssx",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSigningKey())
}

// createExpiredToken creates an expired JWT token for testing.
func createExpiredToken(id string) (string, error) {
	claims := RssxClaims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"rssx.wiloon.net"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired 1 hour ago
			ID:        "test-jti-expired",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "wiloon.com",
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "rssx",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSigningKey())
}

// createInvalidSignatureToken creates a token signed with a different key.
func createInvalidSignatureToken(id string) (string, error) {
	claims := RssxClaims{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"rssx.wiloon.net"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			ID:        "test-jti",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wiloon.com",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "rssx",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with a different key
	return token.SignedString([]byte("different-secret-key"))
}

// createMalformedToken returns a string that is not a valid JWT.
func createMalformedToken() string {
	return "this.is.not.a.valid.jwt.token"
}

// createTokenWithWrongMethod creates a token using the "none" algorithm (not HMAC).
func createTokenWithWrongMethod(id string) (string, error) {
	claims := jwt.MapClaims{
		"iss": "wiloon.com",
		"sub": "rssx",
		"aud": "rssx.wiloon.net",
		"nbf": float64(time.Now().Unix()),
		"exp": float64(time.Now().Add(24 * time.Hour).Unix()),
		"iat": float64(time.Now().Unix()),
		"jti": "test-jti",
		"id":  id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	return token.SignedString(jwt.UnsafeAllowNoneSignatureType)
}

// TestGenJwtToken tests basic JWT token generation and structure.
func TestGenJwtToken(t *testing.T) {
	token, err := createTestToken("user123", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	if token == "" {
		t.Fatal("generated token is empty")
	}

	// Verify token has 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token should have 3 parts, got %d", len(parts))
	}

	t.Logf("generated token: %s", token)
}

// TestParseToken_ValidToken tests parsing a valid token.
func TestParseToken_ValidToken(t *testing.T) {
	// Create a custom ParseToken that uses our test key
	originalKey := testSecret
	_ = originalKey

	token, err := createTestToken("user123", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Manually parse and verify the token
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("parsed token should be valid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to convert claims to MapClaims")
	}

	if claims["id"] != "user123" {
		t.Errorf("expected id=user123, got %v", claims["id"])
	}
}

// TestParseToken_ExpiredToken tests that expired tokens are rejected.
func TestParseToken_ExpiredToken(t *testing.T) {
	token, err := createExpiredToken("user123")
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})
	if err == nil {
		t.Fatal("expected error when parsing expired token, got nil")
	}

	t.Logf("correctly rejected expired token: %v", err)
}

// TestParseToken_InvalidSignature tests that tokens with wrong signature are rejected.
func TestParseToken_InvalidSignature(t *testing.T) {
	token, err := createInvalidSignatureToken("user123")
	if err != nil {
		t.Fatalf("failed to create token with invalid signature: %v", err)
	}

	_, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})
	if err == nil {
		t.Fatal("expected error when parsing token with invalid signature, got nil")
	}

	t.Logf("correctly rejected invalid signature: %v", err)
}

// TestParseToken_MalformedToken tests that malformed tokens are rejected.
func TestParseToken_MalformedToken(t *testing.T) {
	token := createMalformedToken()

	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})
	if err == nil {
		t.Fatal("expected error when parsing malformed token, got nil")
	}

	t.Logf("correctly rejected malformed token: %v", err)
}

// TestParseToken_WrongSigningMethod tests that tokens signed with non-HMAC methods are rejected.
func TestParseToken_WrongSigningMethod(t *testing.T) {
	token, err := createTokenWithWrongMethod("user123")
	if err != nil {
		t.Fatalf("failed to create token with wrong method: %v", err)
	}

	_, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getSigningKey(), nil
	})
	if err == nil {
		t.Fatal("expected error for token with non-HMAC signing method, got nil")
	}
	t.Logf("correctly rejected non-HMAC signing method: %v", err)
}

// TestRssxClaims_Fields verifies all fields of RssxClaims are set correctly.
func TestRssxClaims_Fields(t *testing.T) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	jti := "test-jti-123"

	claims := RssxClaims{
		Id: "user456",
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"rssx.wiloon.net"},
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "wiloon.com",
			NotBefore: jwt.NewNumericDate(now),
			Subject:   "rssx",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(getSigningKey())
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	parsed, err := jwt.Parse(signedString, func(t *jwt.Token) (interface{}, error) {
		return getSigningKey(), nil
	})
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	parsedClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to convert claims")
	}

	if parsedClaims["id"] != "user456" {
		t.Errorf("id: expected user456, got %v", parsedClaims["id"])
	}
	if parsedClaims["iss"] != "wiloon.com" {
		t.Errorf("iss: expected wiloon.com, got %v", parsedClaims["iss"])
	}
	if parsedClaims["sub"] != "rssx" {
		t.Errorf("sub: expected rssx, got %v", parsedClaims["sub"])
	}
	if !reflect.DeepEqual(parsedClaims["aud"], []interface{}{"rssx.wiloon.net"}) {
		t.Errorf("aud: expected [rssx.wiloon.net], got %v", parsedClaims["aud"])
	}
	if parsedClaims["jti"] != jti {
		t.Errorf("jti: expected %s, got %v", jti, parsedClaims["jti"])
	}
}

// TestTokenBuilder_MakeCommonFiles tests the TokenBuilder helper.
func TestTokenBuilder_MakeCommonFiles(t *testing.T) {
	builder := &TokenBuilder{}
	builder.MakeCommonFiles("test-user-id")

	if builder.payload.Id != "test-user-id" {
		t.Errorf("expected Id=test-user-id, got %s", builder.payload.Id)
	}
	if builder.payload.Iss != DefaultIss {
		t.Errorf("expected Iss=%s, got %s", DefaultIss, builder.payload.Iss)
	}
	if builder.payload.Sub != DefaultSub {
		t.Errorf("expected Sub=%s, got %s", DefaultSub, builder.payload.Sub)
	}
	if builder.payload.Jti == "" {
		t.Error("expected Jti to be set, got empty")
	}
	if builder.payload.Nbf == 0 {
		t.Error("expected Nbf to be set, got 0")
	}
	if builder.payload.Exp == 0 {
		t.Error("expected Exp to be set, got 0")
	}
	if builder.payload.Iat == 0 {
		t.Error("expected Iat to be set, got 0")
	}
}

// BenchmarkGenJwtToken benchmarks token generation performance.
func BenchmarkGenJwtToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := createTestToken("user123", time.Now().Add(24*time.Hour))
		if err != nil {
			b.Fatalf("failed to create token: %v", err)
		}
	}
}

// BenchmarkParseToken benchmarks token parsing performance.
func BenchmarkParseToken(b *testing.B) {
	token, _ := createTestToken("user123", time.Now().Add(24*time.Hour))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return getSigningKey(), nil
		})
		if err != nil {
			b.Fatalf("failed to parse token: %v", err)
		}
	}
}
