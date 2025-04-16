import React, { useState } from 'react';
import axios from 'axios';

// Componente de Login em React
function LoginComponent() {
  // States para armazenar email, senha, erro e token recebido
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [token, setToken] = useState("");

  // Função que é chamada quando o usuário clica no botão de login
  const handleLogin = async () => {
    try {
      // Faz uma requisição POST para o endpoint /login da API
      // O objeto enviado no body é convertido automaticamente para JSON
      const response = await axios.post("http://localhost:8080/login", 
        {
          email: email,
          password: password
        },
        {
          headers: { "Content-Type": "application/json" } // Garante que o payload seja interpretado como JSON
        }
      );

      // Se o login for bem-sucedido, a resposta conterá o token JWT
      setToken(response.data.token);
      setErrorMessage(""); // Limpa a mensagem de erro (se houver)
      console.log("Token recebido:", response.data.token);

      // (Opcional) Salve o token no localStorage se desejar persistência entre sessões:
      // localStorage.setItem("jwt", response.data.token);

    } catch (error) {
      // Em caso de erro, loga o erro no console para debug e seta a mensagem para exibição na UI
      console.error("Error during login:", error);
      setErrorMessage("Falha no login. Verifique suas credenciais e tente novamente.");
    }
  };

  return (
    <div style={{ padding: "20px", fontFamily: "Arial" }}>
      <h2>Login</h2>

      {/* Campo para inserir o e-mail */}
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={e => setEmail(e.target.value)}
        style={{ marginRight: "10px", padding: "8px" }}
      />

      {/* Campo para inserir a senha */}
      <input
        type="password"
        placeholder="Senha"
        value={password}
        onChange={e => setPassword(e.target.value)}
        style={{ marginRight: "10px", padding: "8px" }}
      />

      {/* Botão de login */}
      <button onClick={handleLogin} style={{ padding: "8px 16px" }}>
        Entrar
      </button>

      {/* Exibe a mensagem de erro (em vermelho) se houver falha no login */}
      {errorMessage && <p style={{ color: 'red', marginTop: "10px" }}>{errorMessage}</p>}

      {/* (Opcional) Exibe o token JWT recebido para fins de debug */}
      {token && (
        <div style={{ marginTop: "20px" }}>
          <h3>Token JWT:</h3>
          <pre>{token}</pre>
        </div>
      )}
    </div>
  );
}

export default LoginComponent;
