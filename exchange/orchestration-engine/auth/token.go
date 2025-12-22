package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetConsumerJwtFromToken(env string, r *http.Request) (*ConsumerAssertion, error) {
	if env == "local" {
		// Return dummy values in local environment
		return &ConsumerAssertion{
			ClientId:   "passport-app",
			Subscriber: "passport-app",
			Iss:        "https://idp.example.com",
			Aud:        []string{"https://api.example.com"},
			Exp:        time.Now().Add(time.Hour).Unix(),
			Iat:        time.Now().Unix(),
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

	// Parse without verification (signature validation should be done by gateway or verified here if needed)
	// IMPORTANT: In a real scenario, we MUST verify the signature.
	// Assuming the gateway does it, or we need to configure keys for verification.
	// For this task, we follow existing pattern but add claim validation.
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Validate generic claims

	// exp
	if exp, ok := claims[ClaimExp].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token has expired")
		}
	}

	// nbf
	if nbf, ok := claims[ClaimNbf].(float64); ok {
		if time.Now().Unix() < int64(nbf) {
			return nil, fmt.Errorf("token is not valid yet")
		}
	}

	// client_id is required
	clientId, ok := claims[ClaimClientId].(string)
	if !ok || clientId == "" {
		return nil, fmt.Errorf("missing or invalid client_id claim")
	}

	// sub or azp
	subscriber, ok := claims[ClaimSub].(string)
	if !ok || subscriber == "" {
		// fallback to azp if sub is missing
		if azp, ok := claims[ClaimAzp].(string); ok {
			subscriber = azp
		}
	}

	iss, _ := claims[ClaimIss].(string)

	// aud can be string or array of strings
	var aud []string
	if audStr, ok := claims[ClaimAud].(string); ok {
		aud = []string{audStr}
	} else if audList, ok := claims[ClaimAud].([]interface{}); ok {
		for _, a := range audList {
			if s, ok := a.(string); ok {
				aud = append(aud, s)
			}
		}
	}

	exp, _ := claims[ClaimExp].(float64)
	iat, _ := claims[ClaimIat].(float64)

	// Map claims into generic struct
	ca := &ConsumerAssertion{
		ClientId:   clientId,
		Subscriber: subscriber,
		Iss:        iss,
		Aud:        aud,
		Exp:        int64(exp),
		Iat:        int64(iat),
	}

	return ca, nil
}
