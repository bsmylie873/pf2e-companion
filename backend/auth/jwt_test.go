package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-access-secret")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	token, err := GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Should expire after ~15 minutes
	expiry := claims.ExpiresAt.Time
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), expiry, 5*time.Second)
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-refresh-secret")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	token, err := GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Should expire after ~7 days
	expiry := claims.ExpiresAt.Time
	assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), expiry, 10*time.Second)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateToken("this.is.garbage")
	require.Error(t, err)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	_, err := ValidateToken("")
	require.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Generate with one secret
	os.Setenv("JWT_SECRET", "secret-one")
	userID := uuid.New()
	token, err := GenerateAccessToken(userID)
	require.NoError(t, err)

	// Validate with different secret
	os.Setenv("JWT_SECRET", "secret-two")
	_, err = ValidateToken(token)
	require.Error(t, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Craft an already-expired token
	userID := uuid.New()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	t2 := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t2.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = ValidateToken(tokenStr)
	require.Error(t, err)
}

func TestGenerateAccessToken_DifferentUsersProduceDifferentTokens(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	u1, u2 := uuid.New(), uuid.New()
	t1, err1 := GenerateAccessToken(u1)
	t2, err2 := GenerateAccessToken(u2)
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, t1, t2)
}
