package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h handlers.URLHandler, Method string) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatal(err)
		}
		defer logger.Sync()

		sugar := *logger.Sugar()

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		switch Method {
		case "GetShortUrlText":
			{
				h.GetHandle(&lw, r)
			}
		case "PostRootText":
			{
				h.PostHandle(&lw, r)
			}
		case "PostApiShortenJson":
			{
				h.PostHandleJSON(&lw, r)
			}
		default:
			{
				http.Error(w, "Not supported", http.StatusBadRequest)
				return
			}
		}

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)

	}
	return http.HandlerFunc(logFn)
}
