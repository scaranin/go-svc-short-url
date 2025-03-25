package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type AuthConfig struct {
	CookieName string
	SecretKey  string
	TokenExp   time.Duration
	UserID     string
}

func NewAuthConfig() AuthConfig {
	return AuthConfig{
		CookieName: "auth_token",
		SecretKey:  "TsoyZhiv",
		TokenExp:   time.Hour,
	}
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (auth *AuthConfig) BuildJWTString() (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(auth.TokenExp)),
		},
		// собственное утверждение
		UserID: uuid.New().String(),
	})

	// создаём строку токена
	authToken, err := token.SignedString([]byte(auth.SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return authToken, nil
}

// FillUserReturnToken возвращает authToken, UserID записывается в auth.UserID
func (auth *AuthConfig) FillUserReturnCookie(incomeCookie *http.Cookie) (*http.Cookie, error) {
	var (
		resAuthToken string
		err          error
	)
	if incomeCookie != nil {
		resAuthToken = incomeCookie.Value
	}

	if len(resAuthToken) == 0 {
		resAuthToken, err = auth.BuildJWTString()
	}
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
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
		Secure:   true,
		Path:     "/",
	}
	// возвращаем ID пользователя в читаемом виде
	return cookie, err
}
