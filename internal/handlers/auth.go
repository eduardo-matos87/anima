package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware é um middleware que protege endpoints que requerem autenticação.
// Ele verifica se o header "Authorization" contém um token JWT válido no formato "Bearer <token>".
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtém o header "Authorization" da requisição.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Divide o header para garantir o formato "Bearer <token>".
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format. Expected 'Bearer <token>'", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Faz o parse do token utilizando as Claims do seu sistema.
		// Note que estamos utilizando a função jwt.ParseWithClaims, passando uma instância de Claims.
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Confirma que o método de assinatura é HMAC (HS256, por exemplo).
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// jwtKey foi definido em user.go; como estamos no mesmo pacote (handlers),
			// podemos usá-lo diretamente.
			return jwtKey, nil
		})
		if err != nil {
			http.Error(w, "Error parsing token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Verifica se o token é válido
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// (Opcional) Você pode extrair as Claims e armazenar informações adicionais no contexto.
		// Exemplo:
		// claims := token.Claims.(*Claims)
		// r.Header.Set("user_id", fmt.Sprintf("%d", claims.ID))

		// Se o token é válido, chama o próximo handler protegido.
		next(w, r)
	}
}
