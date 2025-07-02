package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims represents the custom claims used in the JWT for authentication.
// It embeds the standard `jwt.RegisteredClaims` and adds a UserID to identify
// the subject of the token.
type Claims struct {
	jwt.RegisteredClaims
	// UserID is the unique identifier for the user.
	UserID string
}

// AuthConfig holds the configuration and state for JWT-based authentication.
// It defines the parameters for creating and validating tokens and cookies.
// It also stores the UserID for the currently authenticated user within a request's context.
type AuthConfig struct {
	// CookieName is the name of the HTTP cookie used to store the auth token.
	CookieName string
	// SecretKey is the secret key used to sign and verify JWTs.
	SecretKey string
	// TokenExp is the duration for which a token is valid after being issued.
	TokenExp time.Duration
	// UserID holds the identifier of the user making the current request.
	// It is populated by the authentication middleware/handler after validating a token.
	UserID string
}

// NewAuthConfig creates and returns a new AuthConfig instance with default values.
func NewAuthConfig() AuthConfig {
	return AuthConfig{
		CookieName: "auth_token",
		SecretKey:  "TsoyZhiv",
		TokenExp:   time.Hour,
	}
}

// BuildJWTString creates a new JWT for a new user session.
// It generates a new unique UserID (via UUID), sets the token's expiration time
// based on `auth.TokenExp`, and signs it using the configured `SecretKey`.
// It returns the signed token as a string.
func (auth *AuthConfig) BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auth.TokenExp)),
		},
		UserID: uuid.New().String(),
	})

	authToken, err := token.SignedString([]byte(auth.SecretKey))
	if err != nil {
		return "", err
	}

	return authToken, nil
}

// FillUserReturnCookie is the central function for handling cookie-based JWT authentication.
// It inspects the incoming cookie from the request (`incomeCookie`).
//
// If the cookie is missing or invalid, it generates a new JWT for a new user session
// by calling `BuildJWTString`. If the cookie is valid, it parses the JWT to extract
// the UserID.
//
// In both cases, it populates the `auth.UserID` field with the determined user ID, making
// it available to the rest of the handler chain.
//
// It returns a new `http.Cookie` object (with a refreshed expiration) to be set in the response,
// and an error if one occurred. A special `http.ErrNoCookie` is returned if the
// token could not be validated.
func (auth *AuthConfig) FillUserReturnCookie(incomeCookie *http.Cookie) (*http.Cookie, error) {
	var (
		resAuthToken string
		err          error
	)
	fmt.Println(incomeCookie)
	if incomeCookie != nil {
		resAuthToken = incomeCookie.Value
	}

	if len(resAuthToken) == 0 {
		resAuthToken, err = auth.BuildJWTString()
		if err != nil {
			return nil, err
		}
	}
	claims := &Claims{}
	jwt.ParseWithClaims(resAuthToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(auth.SecretKey), nil
	})

	if len(claims.UserID) == 0 {
		err = http.ErrNoCookie
	} else {
		auth.UserID = claims.UserID
	}
	cookie := &http.Cookie{
		Name:     auth.CookieName,
		Value:    resAuthToken,
		Expires:  time.Now().Add(auth.TokenExp),
		HttpOnly: true,
		//Secure:   true,
		Path: "/",
	}
	return cookie, err
}
