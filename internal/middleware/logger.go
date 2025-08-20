package middleware

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData holds captured response information for logging purposes.
	responseData struct {
		status int
		size   int
	}
	// loggingResponseWriter wraps http.ResponseWriter to capture response details.
	// It implements http.ResponseWriter interface while tracking:
	// - HTTP status code (via WriteHeader)
	// - Response size (via Write)
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Write captures the number of bytes written while delegating to the original ResponseWriter.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the status code while delegating to the original ResponseWriter.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging provides HTTP middleware that logs request/response details.
// Logs include:
//   - Request URI
//   - HTTP method
//   - Response status code
//   - Response duration
//   - Response size in bytes
//
// The middleware uses zap logger in development mode for structured logging.
func WithLogging(h http.Handler) http.Handler {
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

		h.ServeHTTP(&lw, r)

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
