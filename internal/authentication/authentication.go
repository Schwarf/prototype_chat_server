package authentication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Schwarf/prototype_chat_server/internal/models"
)

var jwtKey = []byte("your_secret_key")
var secrets map[string]bool
var registeredClients = make(map[int]models.Client)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func LoadSecrets() {
	filename := "/home/andreas/Documents/chat_secrets/secrets.txt"
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to load secrets: %v", err)
	}

	secrets = make(map[string]bool)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		secret := strings.TrimSpace(line)
		if secret != "" {
			secrets[secret] = true
		}
	}
}

func IsSecretValid(secret string) bool {
	return secrets[secret]
}

func RemoveSecret(secret string) {
	delete(secrets, secret)
}

func IsAlphaNumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func GenerateHash(message string, salt string) string {
	data := message + salt
	hash := sha256.New()
	hash.Write([]byte(data))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func RegisterClient(clientID int, client models.Client) {
	registeredClients[clientID] = client
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
