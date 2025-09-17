package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

type AuthService struct {
	consumerService *ConsumerService
	secretKey       []byte
	tokenExpiry     time.Duration
}

func NewAuthService(consumerService *ConsumerService) *AuthService {
	// Generate a secret key for JWT signing (in production, this should be from config)
	secretKey := make([]byte, 32)
	rand.Read(secretKey)

	return &AuthService{
		consumerService: consumerService,
		secretKey:       secretKey,
		tokenExpiry:     24 * time.Hour, // 24 hours
	}
}

// AuthenticateConsumer authenticates a consumer using consumer_id and secret
func (s *AuthService) AuthenticateConsumer(req models.AuthRequest) (*models.AuthResponse, error) {
	// Get the consumer application by consumer ID
	apps, err := s.consumerService.GetConsumerAppsByConsumerID(req.ConsumerID)
	if err != nil {
		return nil, fmt.Errorf("consumer not found")
	}

	// Find an approved application with credentials
	var credentials *models.Credentials
	for _, app := range apps {
		if app.Status == models.StatusApproved && app.Credentials != nil {
			credentials = app.Credentials
			break
		}
	}

	if credentials == nil {
		return nil, fmt.Errorf("no approved application found for consumer")
	}

	// Verify the secret matches the API secret
	if credentials.APISecret != req.Secret {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(req.ConsumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	expiresAt := time.Now().Add(s.tokenExpiry)

	slog.Info("Consumer authenticated successfully", "consumerId", req.ConsumerID)

	return &models.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenExpiry.Seconds()),
		ExpiresAt:   expiresAt,
		ConsumerID:  req.ConsumerID,
	}, nil
}

// ValidateToken validates an access token and returns consumer information
func (s *AuthService) ValidateToken(token string) (*models.ValidateTokenResponse, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Parse and validate the token
	claims, err := s.parseToken(token)
	if err != nil {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Invalid token: " + err.Error(),
		}, nil
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt) {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Token has expired",
		}, nil
	}

	// Verify consumer still exists and has approved application
	apps, err := s.consumerService.GetConsumerAppsByConsumerID(claims.ConsumerID)
	if err != nil {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Consumer not found",
		}, nil
	}

	// Check if consumer still has approved application
	hasApprovedApp := false
	for _, app := range apps {
		if app.Status == models.StatusApproved {
			hasApprovedApp = true
			break
		}
	}

	if !hasApprovedApp {
		return &models.ValidateTokenResponse{
			Valid: false,
			Error: "Consumer application no longer approved",
		}, nil
	}

	return &models.ValidateTokenResponse{
		Valid:      true,
		ConsumerID: claims.ConsumerID,
	}, nil
}

// generateAccessToken creates a JWT-like token for the consumer
func (s *AuthService) generateAccessToken(consumerID string) (string, error) {
	now := time.Now()
	claims := models.TokenClaims{
		ConsumerID: consumerID,
		IssuedAt:   now,
		ExpiresAt:  now.Add(s.tokenExpiry),
		Issuer:     "gov-dx-sandbox",
	}

	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := s.createSignature(message)

	// Combine all parts
	token := message + "." + signature

	return token, nil
}

// parseToken parses and validates a JWT-like token
func (s *AuthService) parseToken(token string) (*models.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Verify signature
	message := parts[0] + "." + parts[1]
	expectedSignature := s.createSignature(message)
	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid claims encoding")
	}

	var claims models.TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims format")
	}

	return &claims, nil
}

// createSignature creates HMAC-SHA256 signature
func (s *AuthService) createSignature(message string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// GetConsumerIDFromToken extracts consumer ID from a valid token
func (s *AuthService) GetConsumerIDFromToken(token string) (string, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	claims, err := s.parseToken(token)
	if err != nil {
		return "", err
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt) {
		return "", fmt.Errorf("token has expired")
	}

	return claims.ConsumerID, nil
}
