// Arquivo: anima-frontend/src/App.js

import React, { useState } from 'react';
import api from './api';
import './App.css';
import RegisterComponent from './RegisterComponent';  // importamos o registro
import LoginComponent from './LoginComponent';        // seu componente de login

/**
 * Componente principal que alterna entre Registro e Login
 * e, após autenticação, exibe o painel de criação/listagem de treinos.
 */
function App() {
  // controla qual tela está ativa: 'register', 'login' ou 'dashboard'
  const [stage, setStage] = useState<'register'|'login'|'dashboard'>('login');
  // token para saber se está logado
  const [token, setToken] = useState<string | null>(localStorage.getItem('jwt'));
  // estados de treino (para dashboard)
  const [treinoData, setTreinoData] = useState<any>(null);
  const [treinos, setTreinos] = useState<any[]>([]);

  // Callback quando o login é bem‑sucedido
  const onLoginSuccess = (jwt: string) => {
    localStorage.setItem('jwt', jwt);
    setToken(jwt);
    setStage('dashboard');
  };

  // Função para criar treino (igual antes)
  const createTreino = async () => {
    try {
      const resp = await api.post('/treino/criar', {
        nivel: "iniciante",
        objetivo: "emagrecimento",
        dias: 3,
        divisao: "A",
        exercicios: [1,2,11],
      });
      setTreinoData(resp.data);
    } catch (err) {
      console.error(err);
    }
  };

  // Função para listar treinos
  const fetchTreinos = async () => {
    try {
      const resp = await api.get('/treinos');
      setTreinos(resp.data);
    } catch (err) {
      console.error(err);
    }
  };

  // Se não está autenticado, mostra registro ou login
  if (!token) {
    return (
      <div className="App" style={{ padding: 20, fontFamily: 'Arial' }}>
        <h1>Anima App</h1>
        <nav style={{ marginBottom: 20 }}>
          <button onClick={() => setStage('login')} disabled={stage==='login'}>Login</button>
          <button onClick={() => setStage('register')} disabled={stage==='register'}>Registrar</button>
        </nav>

        {stage === 'login' && <LoginComponent onSuccess={onLoginSuccess} />}
        {stage === 'register' && <RegisterComponent />}
      </div>
    );
  }

  // Se estiver logado, mostra o dashboard de treinos
  return (
    <div className="App" style={{ padding: 20, fontFamily: 'Arial' }}>
      <h1>Painel de Treinos</h1>
      <button onClick={createTreino}>Criar Treino</button>
      {treinoData && <pre>{JSON.stringify(treinoData, null, 2)}</pre>}

      <hr style={{ margin: '20px 0' }} />

      <h2>Seus Treinos</h2>
      <button onClick={fetchTreinos}>Carregar Treinos</button>
      {treinos.map(t => (
        <div key={t.id} style={{ border: '1px solid #ccc', padding: 10, marginTop: 10 }}>
          <strong>{t.divisao} – {t.nivel} / {t.objetivo}</strong><br/>
          Dias: {t.dias}<br/>
          Exercícios: {t.exercicios.join(', ')}
        </div>
      ))}
    </div>
  );
}

export default App;
