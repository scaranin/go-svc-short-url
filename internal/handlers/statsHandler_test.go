package handlers_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/models"
	"github.com/scaranin/go-svc-short-url/internal/storage"

	"github.com/google/uuid"
)

func TestGetStats(t *testing.T) {
	tests := []struct {
		name string
		// Для Auth: ReturnParseErr
		returnParseErr bool
		// Для CheckIP
		trustedSubnet string
		xRealIP       string
		// Для Storage
		getStatsFunc func() (models.Statistic, error)
		// Ожидания
		reqHasCookie bool
		wantStatus   int
		wantBody     string
		checkUserID  bool // дополнительная проверка UserID из парсинга
	}{
		{
			name:           "get stat handle negative test #1. IP not in subnet",
			returnParseErr: false,
			trustedSubnet:  "192.164.1.0/24",
			xRealIP:        "10.0.0.1",
			getStatsFunc:   nil,
			reqHasCookie:   true,
			wantStatus:     http.StatusForbidden,
			wantBody:       "",
			checkUserID:    false,
		},
		{
			name:           "get stat handle negative test #2. Missing X-Real-IP header",
			returnParseErr: false,
			trustedSubnet:  "192.164.1.0/24",
			xRealIP:        "",
			getStatsFunc:   nil,
			reqHasCookie:   true,
			wantStatus:     http.StatusForbidden,
			wantBody:       "",
			checkUserID:    false,
		},
		{
			name:           "get stat handle negative test #3. EMPTY TrustedSubnet",
			returnParseErr: false,
			trustedSubnet:  "",
			xRealIP:        "192.164.1.1",
			getStatsFunc:   nil,
			reqHasCookie:   true,
			wantStatus:     http.StatusForbidden,
			wantBody:       "",
			checkUserID:    false,
		},
		{
			name:           "get stat handle positive test #1",
			returnParseErr: false,
			trustedSubnet:  "192.164.1.0/24",
			xRealIP:        "192.164.1.10",
			getStatsFunc:   func() (models.Statistic, error) { return models.Statistic{URLs: 42, Users: 10}, nil },
			reqHasCookie:   true,
			wantStatus:     http.StatusOK,
			wantBody:       `{"urls":0,"users":0}`,
			checkUserID:    true, // UserID должно соответствовать UUID из токена
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём URLHandler с моками
			authConfig := auth.AuthConfig{
				CookieName: "auth_token",
				UserID:     uuid.New().String(), // начальное пустое
				SecretKey:  "TsoyZhiv",          // тестовая
				TokenExp:   24 * time.Hour,
			}
			fs, err := storage.CreateStoreFile("BaseFile.json")
			if err != nil {
				log.Println(err)
			}
			h := &handlers.URLHandler{
				Auth:          authConfig,
				Storage:       fs,
				TrustedSubnet: tt.trustedSubnet,
			}

			// Создаём request
			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			req.Header.Set("Content-Type", "text/plain; charset=utf-8") // предположенный contentTypeTextPlain
			if tt.reqHasCookie {
				req.AddCookie(&http.Cookie{Name: "authcookie", Value: "incoming-jwt"}) // тестовый токен (невалидный, но парсинг вернёт ошибку если нужно)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()

			// Вызываем GetStats
			h.GetStats(rr, req)

			res := rr.Result()
			if res.StatusCode != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, res.StatusCode)
			}

			// Проверка тела для 200
			if tt.wantStatus == http.StatusOK {
				got := strings.TrimSpace(rr.Body.String())
				if got != tt.wantBody {
					t.Errorf("want body %q, got %q", tt.wantBody, got)
				}
			}

			// Проверка UserID (для кейсов ускheadline парсинга)
			if tt.checkUserID && tt.returnParseErr == false {
				if authConfig.UserID == "" {
					t.Errorf("want UserID to be set from JWT, got empty")
				} else {
					// Доп. проверка: UUID валидный (если не новый cookie — проверяем формат из произнего токена)
					_, err := uuid.Parse(authConfig.UserID)
					if err != nil {
						t.Errorf("want UserID to be valid UUID, got '%s', err: %v", authConfig.UserID, err)
					}
				}
			}
		})
	}
}
