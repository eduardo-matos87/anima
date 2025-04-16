package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	// Pacote para gerar e validar JWT
	"github.com/golang-jwt/jwt/v4"
	// Pacote para hash de senha (bcrypt)
	"golang.org/x/crypto/bcrypt"
)

// Chave secreta para assinatura do token JWT.
// Em produção, guarde essa chave em variável de ambiente e use um valor forte.
var jwtKey = []byte("minha_chave_secreta_muito_forte")

// User representa a estrutura de um usuário no sistema.
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"` // não expõe a senha no JSON
}

// Credentials representa os dados que o usuário envia para registro ou login.
type Credentials struct {
	Name     string `json:"name,omitempty"`     // usado somente no registro
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims define as informações que serão incluídas no token JWT.
type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUser trata o cadastro de novo usuário (endpoint POST /register).
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		// Decodifica o JSON enviado pelo cliente
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Gera o hash da senha utilizando bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erro ao processar senha", http.StatusInternalServerError)
			return
		}

		// Insere o novo usuário na tabela 'users'
		result, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
			creds.Name, creds.Email, hashedPassword)
		if err != nil {
			log.Println("Erro ao inserir usuário:", err)
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}

		userID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}

		// Retorna resposta de sucesso
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"mensagem": "Usuário registrado com sucesso",
			"user_id":  userID,
		})
	}
}

// LoginUser trata o login do usuário (endpoint POST /login).
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		// Decodifica o JSON do login
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Busca no banco o usuário com o e-mail fornecido
		var user User
		row := db.QueryRow("SELECT id, name, email, password FROM users WHERE email = ?", creds.Email)
		err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		// Compara a senha enviada com a senha armazenada (hash)
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}

		// Define a expiração do token (ex: 24 horas)
		expirationTime := time.Now().Add(24 * time.Hour)

		// Cria as claims para o token
		claims := &Claims{
			ID:    user.ID,
			Email: user.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		// Cria o token com as claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
			return
		}

		// Retorna o token JWT no JSON de resposta
		json.NewEncoder(w).Encode(map[string]any{
			"token": tokenString,
		})
	}
}
