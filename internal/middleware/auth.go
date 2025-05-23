package middleware

import "net/http"

// AuthMiddleware checks for a valid authentication token
func Auth(next http.HandlerFunc, authToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if no token is configured
		if authToken == "" {
			next(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if token != authToken {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
