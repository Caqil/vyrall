package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	config *config.JWTConfig
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
func NewJWTService(config *config.JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateToken creates a new JWT token for a user
func (s *JWTService) GenerateToken(userID primitive.ObjectID) (string, error) {
	// Set claims
	expirationTime := time.Now().Add(time.Duration(s.config.ExpirationHours) * time.Hour)
	claims := JWTClaims{
		UserID: userID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   userID.Hex(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	signedToken, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", errors.Wrap(err, "Failed to sign JWT token")
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *JWTService) ValidateToken(tokenString string) (*primitive.ObjectID, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, errors.New(errors.CodeInvalidToken, "Invalid token: "+err.Error())
	}

	// Validate claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Convert user ID string to ObjectID
		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			return nil, errors.New(errors.CodeInvalidToken, "Invalid user ID in token")
		}
		return &userID, nil
	}

	return nil, errors.New(errors.CodeInvalidToken, "Invalid token claims")
}
