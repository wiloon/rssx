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

// ---------------------------------------------------------------------------
// Business-layer tests — cover PR-changed code in jwt.go
// ---------------------------------------------------------------------------

// TestNewToken_Claims verifies that NewToken produces a well-formed JWT containing
// all required claims using jwt.RegisteredClaims (jwt/v5 change from StandardClaims).
func TestNewToken_Claims(t *testing.T) {
	const key = "test-rssx-security-key"
	t.Setenv("RSSX_SECURITY_KEY", key)

	token := NewToken("user-id-xyz")
	if token == "" {
		t.Fatal("expected non-empty token string")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("token is malformed: %v", err)
	}

	mc := parsed.Claims.(jwt.MapClaims)

	if mc["id"] != "user-id-xyz" {
		t.Errorf("id: expected user-id-xyz, got %v", mc["id"])
	}
	if mc["iss"] != "wiloon.com" {
		t.Errorf("iss: expected wiloon.com, got %v", mc["iss"])
	}
	if mc["sub"] != "rssx" {
		t.Errorf("sub: expected rssx, got %v", mc["sub"])
	}
	if mc["jti"] == nil || mc["jti"] == "" {
		t.Error("jti must be set and non-empty")
	}
	if mc["exp"] == nil {
		t.Error("exp claim must be set")
	}
	// MapClaims decodes JSON arrays as []interface{}
	aud, ok := mc["aud"].([]interface{})
	if !ok || len(aud) != 1 || aud[0] != "rssx.wiloon.net" {
		t.Errorf("aud: expected [rssx.wiloon.net], got %v (type %T)", mc["aud"], mc["aud"])
	}
	// Header must declare HS256 — confirms RegisteredClaims path is taken
	if parsed.Method.Alg() != "HS256" {
		t.Errorf("alg: expected HS256, got %s", parsed.Method.Alg())
	}
}

// TestNewToken_UniqueJTI verifies uuid.New() produces a distinct JTI per call.
func TestNewToken_UniqueJTI(t *testing.T) {
	t.Setenv("RSSX_SECURITY_KEY", "test-rssx-security-key")

	parser := jwt.NewParser()
	getJTI := func(tok string) string {
		parsed, _, _ := parser.ParseUnverified(tok, jwt.MapClaims{})
		return parsed.Claims.(jwt.MapClaims)["jti"].(string)
	}

	jti1 := getJTI(NewToken("userA"))
	jti2 := getJTI(NewToken("userB"))
	jti3 := getJTI(NewToken("userA")) // same user, different call

	if jti1 == jti2 || jti1 == jti3 || jti2 == jti3 {
		t.Error("expected a unique JTI for every NewToken call")
	}
}

// TestParseToken_ValidRoundtrip tests the NewToken → ParseToken roundtrip via the
// business layer (zero coverage before this PR's test additions).
func TestParseToken_ValidRoundtrip(t *testing.T) {
	const key = "rssx-roundtrip-key"
	t.Setenv("RSSX_SECURITY_KEY", key)

	const userID = "roundtrip-user-123"
	token := NewToken(userID)

	payload, err := ParseToken(token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if payload.Id != userID {
		t.Errorf("Id: expected %s, got %s", userID, payload.Id)
	}
	if payload.Iss != "wiloon.com" {
		t.Errorf("Iss: expected wiloon.com, got %s", payload.Iss)
	}
	if payload.Sub != "rssx" {
		t.Errorf("Sub: expected rssx, got %s", payload.Sub)
	}
	if payload.Exp == 0 {
		t.Error("Exp must be non-zero")
	}
	if payload.Iat == 0 {
		t.Error("Iat must be non-zero")
	}
	if payload.Jti == "" {
		t.Error("Jti must be set")
	}
}

// TestParseToken_AudNotMappedToSub is a regression test for the PR fix.
// Before the fix: jwtPayload.Sub = claims["aud"].(string)  — wrong field + potential panic.
// After the fix:  Sub holds the "sub" claim, Aud holds the "aud" claim (safe assertion).
func TestParseToken_AudNotMappedToSub(t *testing.T) {
	const key = "aud-regression-key"
	t.Setenv("RSSX_SECURITY_KEY", key)

	payload, err := ParseToken(NewToken("aud-test-user"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Sub must contain "rssx" (the "sub" claim), not "rssx.wiloon.net" (the "aud" value).
	if payload.Sub == "rssx.wiloon.net" {
		t.Error("regression: aud value has been stored in Sub — PR bug fix has regressed")
	}
	if payload.Sub != "rssx" {
		t.Errorf("Sub: expected rssx, got %s", payload.Sub)
	}
}

// TestParseToken_BusinessLayer_Expired verifies the business-layer ParseToken
// rejects expired tokens (error-first check added in the PR).
func TestParseToken_BusinessLayer_Expired(t *testing.T) {
	const key = "expired-test-key"
	t.Setenv("RSSX_SECURITY_KEY", key)

	expiredClaims := jwt.MapClaims{
		"iss": "wiloon.com",
		"sub": "rssx",
		"exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
		"iat": float64(time.Now().Add(-2 * time.Hour).Unix()),
		"jti": "expired-jti",
		"id":  "some-user",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	signed, _ := tok.SignedString([]byte(key))

	_, err := ParseToken(signed)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// TestParseToken_BusinessLayer_InvalidSignature verifies the business-layer ParseToken
// rejects tokens signed with a different key.
func TestParseToken_BusinessLayer_InvalidSignature(t *testing.T) {
	const key = "correct-key"
	t.Setenv("RSSX_SECURITY_KEY", key)

	claims := jwt.MapClaims{
		"iss": "wiloon.com",
		"sub": "rssx",
		"exp": float64(time.Now().Add(24 * time.Hour).Unix()),
		"id":  "some-user",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte("wrong-key"))

	_, err := ParseToken(signed)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
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
