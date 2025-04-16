// Arquivo: anima/internal/handlers/user.go

package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// jwtKey é a chave secreta usada para assinar os tokens JWT.
// Em produção, carregue-a de uma variável de ambiente.
var jwtKey = []byte("minha_chave_secreta_muito_forte")

// Credentials representa os dados de login ou registro do usuário.
type Credentials struct {
	Name     string `json:"name,omitempty"` // só usado no registro
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User representa o usuário armazenado no banco.
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// Claims representa as claims do JWT.
type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUser cadastra um novo usuário.
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) Lê todo o body para debug
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			logrus.Errorf("RegisterUser: falha ao ler body: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		logrus.Debugf("RegisterUser body raw: %s", string(raw))
		// Reconstrói r.Body para o decoder
		r.Body = io.NopCloser(bytes.NewBuffer(raw))

		// 2) Decodifica JSON
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Errorf("RegisterUser: JSON inválido: %v", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// 3) Gera hash da senha
		hashed, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("RegisterUser: bcrypt error: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		// 4) Insere usuário no banco
		res, err := db.Exec(
			"INSERT INTO users(name, email, password) VALUES(?,?,?)",
			creds.Name, creds.Email, string(hashed),
		)
		if err != nil {
			logrus.Errorf("RegisterUser: DB insert error: %v", err)
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}

		id, _ := res.LastInsertId()
		logrus.Infof("RegisterUser: usuário criado, email=%s, id=%d", creds.Email, id)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Usuário registrado com sucesso",
			"user_id":  id,
		})
	}
}

// LoginUser autentica o usuário e retorna um JWT.
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) Lê e loga o body para debug
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			logrus.Errorf("LoginUser: falha ao ler body: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		logrus.Debugf("LoginUser body raw: %s", string(raw))
		// Reconstrói r.Body para o decoder
		r.Body = io.NopCloser(bytes.NewBuffer(raw))

		// 2) Decodifica JSON com credenciais
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Errorf("LoginUser: JSON inválido: %v", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// 3) Busca usuário no banco
		var user User
		err = db.QueryRow(
			"SELECT id, name, email, password FROM users WHERE email = ?",
			creds.Email,
		).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			logrus.Errorf("LoginUser: usuário não encontrado, email=%s, err=%v", creds.Email, err)
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		// 4) Verifica senha
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			logrus.Errorf("LoginUser: senha incorreta para email=%s", creds.Email)
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}

		// 5) Gera as claims e o token JWT
		expiration := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			ID:    user.ID,
			Email: user.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiration),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(jwtKey)
		if err != nil {
			logrus.Errorf("LoginUser: erro ao gerar token para email=%s: %v", creds.Email, err)
			http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
			return
		}

		logrus.Infof("LoginUser: autenticação bem-sucedida, email=%s", creds.Email)

		// 6) Retorna o token JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": signed,
		})
	}
}
