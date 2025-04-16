// Arquivo: anima-frontend/src/App.js

import React, { useState } from 'react';
import api from './api';    // Instância do Axios configurada em src/api.js
import './App.css';         // Seus estilos globais

/**
 * Componente principal da aplicação Anima Front‑End.
 * Gerencia o login do usuário e a chamada protegida para criar um treino.
 */
function App() {
  // 📧 State para guardar o valor do campo de e‑mail
  const [loginEmail, setLoginEmail] = useState("");
  // 🔒 State para guardar o valor do campo de senha
  const [loginPassword, setLoginPassword] = useState("");
  // ⚠️ State para exibir mensagens de erro na UI
  const [errorMessage, setErrorMessage] = useState("");
  // 📊 State para armazenar a resposta do endpoint /treino/criar
  const [treinoData, setTreinoData] = useState(null);

  /**
   * loginUser() → Envia POST /login com email e senha.
   * Se bem‑sucedido, salva o token JWT no localStorage.
   */
  const loginUser = async () => {
    // 1️⃣ Para debug: veja no console o que está sendo enviado
    console.log("Enviando login:", { email: loginEmail, password: loginPassword });

    try {
      // ⬆️ POST /login { email, password }
      const response = await api.post(
        "/login",
        { email: loginEmail, password: loginPassword },
        { headers: { "Content-Type": "application/json" } }
      );

      // 🔑 Salva o token JWT no localStorage
      localStorage.setItem("jwt", response.data.token);
      setErrorMessage(""); // 🚫 Limpa mensagens de erro anteriores
      alert("Login realizado com sucesso!");
    } catch (error) {
      // 🐞 Log detalhado no console para debug
      console.error("Erro no login:", error);
      // 📝 Exibe mensagem de falha na UI
      setErrorMessage("Falha no login. Verifique suas credenciais e tente novamente.");
    }
  };

  /**
   * createTreino() → Envia POST /treino/criar (endpoint protegido),
   * usa o token JWT do localStorage automaticamente.
   */
  const createTreino = async () => {
    try {
      // ⬆️ POST /treino/criar { nivel, objetivo, dias, divisao, exercicios }
      const response = await api.post("/treino/criar", {
        nivel: "iniciante",
        objetivo: "emagrecimento",
        dias: 3,
        divisao: "A",
        exercicios: [1, 2, 11],
      });
      // 📈 Atualiza o state com a resposta
      setTreinoData(response.data);
      setErrorMessage("");
    } catch (error) {
      console.error("Erro ao criar treino:", error);
      setErrorMessage("Erro ao criar treino. Verifique se está logado e tente novamente.");
    }
  };

  return (
    <div className="App" style={{ padding: 20, fontFamily: 'Arial, sans-serif' }}>
      {/* Título principal */}
      <h1>Anima Front‑End</h1>

      {/* === Seção de Login === */}
      <section style={{ marginBottom: 40 }}>
        <h2>Login</h2>
        {/* Campo de e‑mail */}
        <input
          type="email"
          placeholder="Email"
          value={loginEmail}
          onChange={e => setLoginEmail(e.target.value)}
          style={{ marginRight: 10, padding: 8 }}
        />
        {/* Campo de senha */}
        <input
          type="password"
          placeholder="Senha"
          value={loginPassword}
          onChange={e => setLoginPassword(e.target.value)}
          style={{ marginRight: 10, padding: 8 }}
        />
        {/* Botão de login (type="button" evita submit de form vazio) */}
        <button type="button" onClick={loginUser} style={{ padding: '8px 16px' }}>
          Entrar
        </button>
        {/* Exibe mensagem de erro se login falhar */}
        {errorMessage && (
          <p style={{ color: 'red', marginTop: 10 }}>{errorMessage}</p>
        )}
      </section>

      {/* === Seção de Criação de Treino (Protegida) === */}
      <section>
        <h2>Criar Treino (Requer login)</h2>
        {/* Botão que dispara a criação do treino */}
        <button type="button" onClick={createTreino} style={{ padding: '8px 16px' }}>
          Criar Treino
        </button>
        {/* Exibe o JSON da resposta, se houver */}
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
