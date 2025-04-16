// Arquivo: anima-frontend/src/App.js

import React, { useState } from 'react';
import api from './api';    // Instância do Axios configurada em src/api.js
import './App.css';         // Seus estilos globais

/**
 * Componente principal da aplicação Anima Front‑End.
 * Gerencia o login do usuário, criação de treino protegido e listagem de treinos.
 */
function App() {
  // 📧 State para armazenar o e‑mail digitado
  const [loginEmail, setLoginEmail] = useState("");
  // 🔒 State para armazenar a senha digitada
  const [loginPassword, setLoginPassword] = useState("");
  // ⚠️ State para exibir mensagens de erro (login ou criação de treino)
  const [errorMessage, setErrorMessage] = useState("");
  // 📈 State para armazenar a resposta do endpoint /treino/criar
  const [treinoData, setTreinoData] = useState(null);
  // 📊 State para armazenar a lista de treinos retornada pelo backend
  const [treinos, setTreinos] = useState([]);

  /**
   * loginUser → Faz POST /login com email e senha.
   * Salva o token JWT no localStorage se for bem‑sucedido.
   */
  const loginUser = async () => {
    console.log("🔍 Payload login:", { email: loginEmail, password: loginPassword });
    try {
      const response = await api.post(
        "/login",
        { email: loginEmail, password: loginPassword },
        { headers: { "Content-Type": "application/json" } }
      );
      localStorage.setItem("jwt", response.data.token);
      setErrorMessage("");
      alert("✅ Login realizado com sucesso!");
    } catch (err) {
      console.error("❌ Erro no login:", err);
      setErrorMessage("Falha no login. Verifique suas credenciais.");
    }
  };

  /**
   * createTreino → Faz POST /treino/criar (endpoint protegido).
   * O interceptor de api.js anexa o JWT automaticamente.
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
      setErrorMessage("Erro ao criar treino. Está autenticado?");
    }
  };

  /**
   * fetchTreinos → Faz GET /treinos (endpoint protegido).
   * Atualiza o state com o array de treinos detalhados.
   */
  const fetchTreinos = async () => {
    try {
      const response = await api.get("/treinos");
      setTreinos(response.data);
    } catch (err) {
      console.error("❌ Erro ao listar treinos:", err);
      setErrorMessage("Erro ao carregar treinos. Está autenticado?");
    }
  };

  return (
    <div className="App" style={{ padding: 20, fontFamily: 'Arial, sans-serif' }}>
      <h1>Anima Front‑End</h1>

      {/* === Seção de Login === */}
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
        <button type="button" onClick={loginUser} style={{ padding: '8px 16px' }}>
          Entrar
        </button>
        {errorMessage && (
          <p style={{ color: 'red', marginTop: 10 }}>{errorMessage}</p>
        )}
      </section>

      {/* === Seção de Criação de Treino === */}
      <section style={{ marginBottom: 40 }}>
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

      {/* === Seção de Listagem de Treinos === */}
      <section>
        <h2>Seus Treinos</h2>
        <button type="button" onClick={fetchTreinos} style={{ padding: '8px 16px' }}>
          Carregar Treinos
        </button>
        {treinos.length > 0 && treinos.map(t => (
          <div
            key={t.id}
            style={{
              border: '1px solid #ccc',
              borderRadius: 4,
              padding: 10,
              marginTop: 10
            }}
          >
            <strong>{t.divisao} – {t.nivel} / {t.objetivo}</strong><br/>
            Dias: {t.dias}<br/>
            Exercícios: {t.exercicios.join(', ')}
          </div>
        ))}
      </section>
    </div>
  );
}

export default App;
