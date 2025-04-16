// Arquivo: internal/handlers/user.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// chave secreta para JWT (use env var em produção)
var jwtKey = []byte("minha_chave_secreta_muito_forte")

// Credentials para registro/login
type Credentials struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User model de usuário
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// Claims do JWT
type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUser cadastra novo usuário
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Error("RegisterUser: JSON inválido:", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Error("RegisterUser:", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		res, err := db.Exec("INSERT INTO users(name,email,password) VALUES(?,?,?)",
			creds.Name, creds.Email, string(hash))
		if err != nil {
			logrus.Error("RegisterUser:", err)
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Usuário registrado com sucesso",
			"user_id":  id,
		})
	}
}

// LoginUser autentica usuário e retorna JWT
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Error("LoginUser: JSON inválido:", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		var user User
		err := db.QueryRow("SELECT id,name,email,password FROM users WHERE email=?", creds.Email).
			Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			logrus.Error("LoginUser: usuário não encontrado:", err)
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			logrus.Error("LoginUser: senha incorreta")
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}
		exp := time.Now().Add(24 * time.Hour)
		claims := &Claims{ID: user.ID, Email: user.Email, RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, err := token.SignedString(jwtKey)
		if err != nil {
			logrus.Error("LoginUser: erro ao gerar token:", err)
			http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
			return
		}
		logrus.Info("LoginUser: sucesso:", creds.Email)
		json.NewEncoder(w).Encode(map[string]string{"token": s})
	}
}
