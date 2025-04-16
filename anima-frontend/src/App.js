import React, { useState } from 'react';
import axios from 'axios';
import './App.css';

function App() {
  // States para login e para guardar o token e treino
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [token, setToken] = useState("");
  const [treinoData, setTreinoData] = useState(null);
  const [errorMessage, setErrorMessage] = useState("");

  // Função para logar o usuário
  const loginUser = async () => {
    try {
      const response = await axios.post("http://localhost:8080/login", {
        email: loginEmail,
        password: loginPassword
      });
      setToken(response.data.token);
      setErrorMessage("");
      alert("Login realizado com sucesso!");
    } catch (error) {
      console.error(error);
      setErrorMessage("Falha no login");
    }
  };

  // Função para buscar treino
  const fetchTreino = async () => {
    try {
      const response = await axios.get("http://localhost:8080/treino", {
        params: {
          nivel: "iniciante",
          objetivo: "emagrecimento"
        }
      });
      setTreinoData(response.data);
      setErrorMessage("");
    } catch (error) {
      console.error(error);
      setErrorMessage("Erro ao buscar treino");
    }
  };

  return (
    <div style={{ padding: "20px", fontFamily: "Arial" }}>
      <h1>Anima Front-End</h1>

      <div style={{ marginBottom: "30px" }}>
        <h2>Login</h2>
        <input
          type="email"
          placeholder="Email"
          value={loginEmail}
          onChange={e => setLoginEmail(e.target.value)}
          style={{ marginRight: "10px" }}
        />
        <input
          type="password"
          placeholder="Senha"
          value={loginPassword}
          onChange={e => setLoginPassword(e.target.value)}
          style={{ marginRight: "10px" }}
        />
        <button onClick={loginUser}>Login</button>
        {errorMessage && <p style={{ color: "red" }}>{errorMessage}</p>}
      </div>

      <div style={{ marginBottom: "30px" }}>
        <h2>Buscar Treino</h2>
        <button onClick={fetchTreino}>Buscar Treino</button>
        {treinoData && (
          <div style={{ marginTop: "10px", background: "#f8f8f8", padding: "10px" }}>
            <h3>Treino</h3>
            <pre>{JSON.stringify(treinoData, null, 2)}</pre>
          </div>
        )}
      </div>

      {token && (
        <div>
          <h2>Token JWT</h2>
          <pre>{token}</pre>
        </div>
      )}
    </div>
  );
}

export default App;
