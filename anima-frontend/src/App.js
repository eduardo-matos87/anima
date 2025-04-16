// Arquivo: anima-frontend/src/App.js

import React, { useState } from 'react';
import api from './api';    // Instância do Axios
import './App.css';         // Seus estilos

/**
 * Componente principal da aplicação web Anima Front‑End.
 * Gerencia o login do usuário e a chamada protegida para criar um treino.
 */
function App() {
  // States para os campos de login e para exibir erros/respostas
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [treinoData, setTreinoData] = useState(null);

  /**
   * Envia a requisição de login.
   * Agora com DEBUG no console e botão tipo="button" para evitar recarregamento de página.
   */
  const loginUser = async () => {
    // 1) DEBUG: veja exatamente o que está sendo enviado
    console.log("🔍 Payload de login:", {
      email: loginEmail,
      password: loginPassword
    });

    try {
      const response = await api.post(
        "/login",
        { email: loginEmail, password: loginPassword },
        { headers: { "Content-Type": "application/json" } }
      );
      // 2) Se chegar aqui, salvamos o token
      localStorage.setItem("jwt", response.data.token);
      setErrorMessage("");
      alert("✅ Login realizado com sucesso!");
    } catch (err) {
      console.error("❌ Erro no login:", err);
      setErrorMessage("Falha no login. Verifique suas credenciais.");
    }
  };

  /**
   * Envia a requisição para criar um treino (endpoint protegido).
   * O interceptor de api.js adiciona o token automaticamente.
   */
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
    } catch (err) {
      console.error("❌ Erro ao criar treino:", err);
      setErrorMessage("Erro ao criar treino. Está logado?");
    }
  };

  return (
    <div className="App" style={{ padding: 20, fontFamily: 'Arial, sans-serif' }}>
      <h1>Anima Front‑End</h1>

      {/* === LOGIN === */}
      <section style={{ marginBottom: 40 }}>
        <h2>Login</h2>
        <input
          type="email"
          placeholder="Email"
          value={loginEmail}
          onChange={e => setLoginEmail(e.target.value)}
          style={{ marginRight: 10, padding: 8 }}
        />
        <input
          type="password"
          placeholder="Senha"
          value={loginPassword}
          onChange={e => setLoginPassword(e.target.value)}
          style={{ marginRight: 10, padding: 8 }}
        />
        {/* Botão type="button" evita comportamento de submit padrão */}
        <button type="button" onClick={loginUser} style={{ padding: '8px 16px' }}>
          Entrar
        </button>
        {errorMessage && (
          <p style={{ color: 'red', marginTop: 10 }}>{errorMessage}</p>
        )}
      </section>

      {/* === CRIAR TREINO === */}
      <section>
        <h2>Criar Treino (Requer login)</h2>
        <button type="button" onClick={createTreino} style={{ padding: '8px 16px' }}>
          Criar Treino
        </button>
        {treinoData && (
          <pre style={{ background: '#f4f4f4', padding: 10, marginTop: 20 }}>
            {JSON.stringify(treinoData, null, 2)}
          </pre>
        )}
      </section>
    </div>
  );
}

export default App;
