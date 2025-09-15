package grpc

import (
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/scaranin/go-svc-short-url/internal/auth"
)

type authService struct {
	auth *auth.AuthConfig
}

func NewAuthService(authConfig *auth.AuthConfig) AuthService {
	return &authService{auth: authConfig}
}

func (a *authService) GetUserIDFromCookie(cookieHeader string) (string, error) {
	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, a.auth.CookieName+"=") {
			token := strings.TrimPrefix(cookie, a.auth.CookieName+"=")

			claims := &auth.Claims{}
			_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(a.auth.SecretKey), nil
			})

			if err != nil {
				return "", err
			}

			return claims.UserID, nil
		}
	}
	return "", nil
}

func (a *authService) CookieName() string {
	return a.auth.CookieName
}
