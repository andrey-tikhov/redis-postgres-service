package validation

import (
	"fmt"
	"net/http"
	"redis-postgres-service/entity"
)

const (
	nilRequest = "nil request received."
)

// httpMethodCheckBuilder is a helper constructor to build a middleware that check
// if the incoming request has a correct http method
func httpMethodCheckBuilder(method string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, fmt.Sprintf(entity.MethodNotAllowed, r.Method), http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var HttpPostCheck = httpMethodCheckBuilder(http.MethodPost)

// NotNilRequest is a middleware that blocks nil requests from going through
func NotNilRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r == nil {
			http.Error(w, fmt.Sprintf(entity.BadRequest, nilRequest), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
