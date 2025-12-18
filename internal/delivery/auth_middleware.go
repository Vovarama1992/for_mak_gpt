package delivery

import (
	"net/http"
	"strings"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

func AuthMiddleware(auth ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(h, "Bearer ")
			ok, err := auth.ValidateToken(r.Context(), token)
			if err != nil || !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
