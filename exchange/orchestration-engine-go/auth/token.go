package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type ConsumerAssertion struct {
	ApplicationUuid string
	Subscriber      string
	ApplicationId   string
}

const (
	WSO2ClaimPrefix = "http://wso2.org/claims/"

	ClaimSubscriber = WSO2ClaimPrefix + "subscriber"
	ClaimAppUUID    = WSO2ClaimPrefix + "applicationUUId"
	ClaimAppID      = WSO2ClaimPrefix + "applicationid"
)

func GetConsumerJwtFromToken(env string, r *http.Request) (*ConsumerAssertion, error) {
	if env == "local" {
		// Return dummy values in local environment
		return &ConsumerAssertion{
			ApplicationUuid: "passport-app",
			Subscriber:      "passport-app",
			ApplicationId:   "passport-app",
		}, nil
	}

	// Check for token in X-JWT-Assertion header first, then Authorization header
	tokenString := r.Header.Get("X-JWT-Assertion")
	if tokenString == "" {
		// Fallback to standard Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return nil, fmt.Errorf("missing token")
		}

		// Remove "Bearer " prefix if present
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			tokenString = authHeader
		}
	}

	// Parse without validation (only safe if API gateway already validated it!)
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Map claims into struct
	ca := &ConsumerAssertion{
		ApplicationUuid: fmt.Sprintf("%v", claims[ClaimAppUUID]),
		Subscriber:      fmt.Sprintf("%v", claims[ClaimSubscriber]),
		ApplicationId:   fmt.Sprintf("%v", claims[ClaimAppID]),
	}

	return ca, nil
}
