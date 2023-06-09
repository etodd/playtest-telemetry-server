package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type middleware func(http.Handler) http.Handler

type middlewares []middleware

func (mws middlewares) WrapFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return mws.Wrap(http.HandlerFunc(f))
}

func (mws middlewares) Wrap(f http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		f = mws[i](f)
	}
	return f
}

func middlewareStack(mws ...middleware) middlewares {
	return mws
}

func basicAuthMiddleware(username, password string) middleware {
	passwordBytes := []byte(password)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok || user != username || subtle.ConstantTimeCompare([]byte(pass), passwordBytes) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerAuthMiddleware(apiKey string) middleware {
	apiKeyBytes := []byte(apiKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(key), apiKeyBytes) != 1 {
				httpError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var errPayloadTooLarge = fmt.Errorf("payload size exceeds the limit")

func limitPayloadMiddleware(maxSize int64) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			defer r.Body.Close()

			if r.ContentLength > maxSize {
				httpError(w, http.StatusRequestEntityTooLarge, errPayloadTooLarge)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var errMethodNotAllowed = fmt.Errorf("method not allowed")

func requireMethod(method string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				httpError(w, http.StatusMethodNotAllowed, errMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

func logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &statusWriter{
			ResponseWriter: w,
		}
		handler.ServeHTTP(writer, r)
		log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, writer.status)
	})
}
