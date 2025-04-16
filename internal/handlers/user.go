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

// jwtKey é a chave secreta para assinatura do token JWT.
// Em produção, guarde este valor em uma variável de ambiente com um valor forte.
var jwtKey = []byte("minha_chave_secreta_muito_forte")

// Credentials representa os dados de login e registro do usuário.
type Credentials struct {
	// Name é utilizado somente no registro.
	Name string `json:"name,omitempty"`
	// Email do usuário para cadastro ou login.
	Email string `json:"email"`
	// Password é a senha do usuário.
	Password string `json:"password"`
}

// User representa a estrutura de um usuário armazenado no banco.
// A propriedade Password não é exposta no JSON.
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// Claims define as informações que serão incluídas no token JWT.
type Claims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUser trata o cadastro de um novo usuário.
// Ele espera um JSON com Name, Email e Password e insere o usuário no banco.
func RegisterUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials

		// Decodifica os dados enviados no corpo da requisição.
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Error("Erro ao decodificar JSON para registro: ", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Gera o hash da senha usando bcrypt.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Error("Erro ao gerar hash da senha: ", err)
			http.Error(w, "Erro ao processar a senha", http.StatusInternalServerError)
			return
		}

		// Insere o novo usuário na tabela "users".
		result, err := db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
			creds.Name, creds.Email, string(hashedPassword))
		if err != nil {
			logrus.Error("Erro ao inserir usuário no banco: ", err)
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}

		// Obtém o ID do usuário recém-criado.
		userID, err := result.LastInsertId()
		if err != nil {
			logrus.Error("Erro ao obter ID do usuário: ", err)
			http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
			return
		}

		// Retorna o resultado do registro.
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Usuário registrado com sucesso",
			"user_id":  userID,
		})
	}
}

// LoginUser autentica o usuário com base em Email e Password.
// Se os dados estiverem corretos, gera e retorna um token JWT válido por 24 horas.
func LoginUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials

		// Decodifica o JSON enviado na requisição de login.
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logrus.Error("Erro ao decodificar JSON para login: ", err)
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Busca o usuário pelo e-mail no banco de dados.
		var user User
		row := db.QueryRow("SELECT id, name, email, password FROM users WHERE email = ?", creds.Email)
		err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			logrus.Error("Usuário não encontrado para o e-mail ", creds.Email, ": ", err)
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		// Compara a senha fornecida com o hash armazenado.
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
			logrus.Error("Senha incorreta para o usuário ", creds.Email)
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}

		// Define a expiração do token para 24 horas.
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			ID:    user.ID,
			Email: user.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		// Cria o token JWT com as claims definidas.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			logrus.Error("Erro ao gerar token JWT para o usuário ", creds.Email, ": ", err)
			http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
			return
		}

		// Registra a autenticação bem-sucedida no log.
		logrus.Info("Usuário autenticado com sucesso: ", creds.Email)

		// Retorna o token JWT na resposta.
		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})
	}
}
