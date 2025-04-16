// src/App.js
import React, { useState } from 'react';
import api from './api';
import './App.css';

function App() {
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [treinoData, setTreinoData] = useState(null);

  // Função para efetuar o login e salvar o token
  const loginUser = async () => {
    try {
      const response = await api.post("/login", {
        email: loginEmail,
        password: loginPassword,
      });
      // Salva o token no localStorage para uso futuro
      localStorage.setItem("jwt", response.data.token);
      setErrorMessage("");
      alert("Login realizado com sucesso!");
    } catch (error) {
      console.error(error);
      setErrorMessage("Falha no login");
    }
  };

  // Função para buscar dados protegidos (exemplo: criação de treino)
  const createTreino = async () => {
    try {
      const response = await api.post("/treino/criar", {
        nivel: "iniciante",
        objetivo: "emagrecimento",
        dias: 3,
        divisao: "A",
        exercicios: [1, 2, 11],
      });
      setTreinoData(response.data);
      setErrorMessage("");
    } catch (error) {
      console.error(error);
      setErrorMessage("Erro ao criar treino");
    }
  };

  return (
    <div className="App">
      <h1>Anima Front-End</h1>

      <div>
        <h2>Login</h2>
        <input
          type="email"
          placeholder="Email"
          value={loginEmail}
          onChange={(e) => setLoginEmail(e.target.value)}
        />
        <input
          type="password"
          placeholder="Senha"
          value={loginPassword}
          onChange={(e) => setLoginPassword(e.target.value)}
        />
        <button onClick={loginUser}>Login</button>
        {errorMessage && <p style={{ color: 'red' }}>{errorMessage}</p>}
      </div>

      <div style={{ marginTop: '20px' }}>
        <h2>Criar Treino (Endpoint Protegido)</h2>
        <button onClick={createTreino}>Criar Treino</button>
        {treinoData && (
          <pre>{JSON.stringify(treinoData, null, 2)}</pre>
        )}
      </div>
    </div>
  );
}

export default App;
