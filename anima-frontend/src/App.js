// Arquivo: anima-frontend/src/App.js

import React, { useState } from 'react';
import api from './api';
import './App.css';
import RegisterComponent from './RegisterComponent';
import LoginComponent from './LoginComponent';

/**
 * Componente principal que alterna entre Registro, Login e Dashboard de treinos.
 */
function App() {
  // Qual tela está ativa: 'register', 'login' ou 'dashboard'
  const [stage, setStage] = useState('login');
  // Token JWT para saber se o usuário está autenticado
  const [token, setToken] = useState(localStorage.getItem('jwt'));
  // Estado para mostrar o treino recém-criado
  const [treinoData, setTreinoData] = useState(null);
  // Lista de treinos carregados do servidor
  const [treinos, setTreinos] = useState([]);

  // Chamado pelo LoginComponent quando o login for bem‑sucedido
  const onLoginSuccess = (jwt) => {
    localStorage.setItem('jwt', jwt);
    setToken(jwt);
    setStage('dashboard');
  };

  // Cria um treino usando o endpoint protegido
  const createTreino = async () => {
    try {
      const resp = await api.post('/treino/criar', {
        nivel: 'iniciante',
        objetivo: 'emagrecimento',
        dias: 3,
        divisao: 'A',
        exercicios: [1, 2, 11],
      });
      setTreinoData(resp.data);
    } catch (err) {
      console.error('Erro ao criar treino:', err);
    }
  };

  // Carrega todos os treinos do usuário
  const fetchTreinos = async () => {
    try {
      const resp = await api.get('/treinos');
      setTreinos(resp.data);
    } catch (err) {
      console.error('Erro ao listar treinos:', err);
    }
  };

  // Se não tiver token, mostra Registro ou Login
  if (!token) {
    return (
      <div className="App" style={{ padding: 20, fontFamily: 'Arial, sans-serif' }}>
        <h1>Anima App</h1>
        <nav style={{ marginBottom: 20 }}>
          <button onClick={() => setStage('login')} disabled={stage === 'login'}>
            Login
          </button>
          <button onClick={() => setStage('register')} disabled={stage === 'register'}>
            Registrar
          </button>
        </nav>

        {stage === 'login' && <LoginComponent onSuccess={onLoginSuccess} />}
        {stage === 'register' && <RegisterComponent />}
      </div>
    );
  }

  // Se tiver token, mostra o dashboard de treinos
  return (
    <div className="App" style={{ padding: 20, fontFamily: 'Arial, sans-serif' }}>
      <h1>Painel de Treinos</h1>

      <section style={{ marginBottom: 40 }}>
        <h2>Criar Treino</h2>
        <button onClick={createTreino}>Criar Treino</button>
        {treinoData && (
          <pre style={{ background: '#f4f4f4', padding: 10, marginTop: 20 }}>
            {JSON.stringify(treinoData, null, 2)}
          </pre>
        )}
      </section>

      <section>
        <h2>Seus Treinos</h2>
        <button onClick={fetchTreinos}>Carregar Treinos</button>
        {treinos.map((t) => (
          <div
            key={t.id}
            style={{
              border: '1px solid #ccc',
              borderRadius: 4,
              padding: 10,
              marginTop: 10,
            }}
          >
            <strong>
              {t.divisao} – {t.nivel} / {t.objetivo}
            </strong>
            <br />
            Dias: {t.dias}
            <br />
            Exercícios: {t.exercicios.join(', ')}
          </div>
        ))}
      </section>
    </div>
  );
}

export default App;
