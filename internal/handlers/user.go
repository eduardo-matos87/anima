package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// jwtKey é a chave secreta para assinar os tokens JWT.
// Em produção, deve ser armazenada em variável de ambiente com um valor seguro.
var jwtKey = []byte("minha_chave_secreta_muito_forte")

// User representa a estrutura de um usuário no sistema.
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"` // Não expõe a senha no JSON
}

// Credentials representa os dados usados para registrar ou fazer login do usuário.
type Credentials struct {
	Name     string `json:"name,omitempty"` // Utilizado somente no registro
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims define as informações que serão incluídas no payload do token JWT.
type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUser cadastra um novo usuário no sistema.
// @Summary Registra um novo usuário
// @Description Cria um novo usuário com nome, email e senha.
// @Tags User
// @Accept json
// @Produce json
// @Param user body Credentials true "Dados de registro"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /register [post]
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		// Decodifica o JSON enviado pelo cliente
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Gera o hash da senha usando bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erro ao processar a senha", http.StatusInternalServerError)
			return
		}

		// Insere o novo usuário na tabela 'users'
		result, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", creds.Name, creds.Email, hashedPassword)
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

		// Retorna a resposta de sucesso
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Usuário registrado com sucesso",
			"user_id":  userID,
		})
	}
}

// LoginUser autentica o usuário e retorna um token JWT.
// @Summary Efetua login do usuário
// @Description Autentica o usuário e retorna um token JWT para acesso à API.
// @Tags User
// @Accept json
// @Produce json
// @Param credentials body Credentials true "Dados de login"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		// Decodifica o JSON do login
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Busca o usuário pelo email
		var user User
		row := db.QueryRow("SELECT id, name, email, password FROM users WHERE email = ?", creds.Email)
		err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		// Compara a senha enviada com o hash armazenado
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}

		// Define o tempo de expiração do token (24 horas)
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			ID:    user.ID,
			Email: user.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		// Cria o token JWT com as claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
			return
		}

		// Retorna o token na resposta
		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})
	}
}
