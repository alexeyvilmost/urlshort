package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return ""
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return ""
	}

	return claims.UserID
}

func WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		jwtAuth, err := r.Cookie("jwt_auth")
		if err != nil {
			if err == http.ErrNoCookie {
				token, err := BuildJWTString(uuid.NewString())
				if err != nil {
					http.Error(w, "Unexpected error while building JWT string", http.StatusInternalServerError)
					return
				}
				cookie := &http.Cookie{
					Path:   "/",
					Name:   "jwt_auth",
					Value:  token,
					MaxAge: 300,
				}
				r.Header.Set("user-id-auth", GetUserID(token))
				r.Header.Set("is-new-user", "true")
				http.SetCookie(w, cookie)
				log.Info().Str("user_id", GetUserID(token)).Msg("")
				h.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unexpected error while getting auth cookie", http.StatusInternalServerError)
			return
		}
		token = jwtAuth.Value

		r.Header.Set("user-id-auth", GetUserID(token))
		log.Info().Str("user_id", GetUserID(token)).Msg("")
		h.ServeHTTP(w, r)
	})
}
