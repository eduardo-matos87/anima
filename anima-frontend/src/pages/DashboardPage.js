// Arquivo: anima-frontend/src/pages/DashboardPage.js

import React, { useEffect, useState } from 'react';
import api from '../api';
import { useHistory, Link } from 'react-router-dom';

export default function DashboardPage() {
  const [treinos, setTreinos] = useState([]);
  const history = useHistory();

  // Ao montar, busca todos os treinos; se falhar, volta ao login
  useEffect(() => {
    const fetchTreinos = async () => {
      try {
        const resp = await api.get('/treinos');
        setTreinos(resp.data);
      } catch {
        history.push('/login');
      }
    };
    fetchTreinos();
  }, [history]);

  // Cria um treino e recarrega a lista
  const createTreino = async () => {
    try {
      await api.post('/treino/criar', {
        nivel: 'iniciante',
        objetivo: 'emagrecimento',
        dias: 3,
        divisao: 'A',
        exercicios: [1, 2, 11],
      });
      const resp = await api.get('/treinos');
      setTreinos(resp.data);
    } catch {
      alert('Erro ao criar treino');
    }
  };

  // Logout: limpa token e volta para login
  const handleLogout = () => {
    localStorage.removeItem('jwt');
    history.push('/login');
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Painel de Treinos</h1>
      <div style={{ float: 'right' }}>
        <Link to="/logout">Sair</Link>
      </div>

      <button onClick={createTreino}>Criar Treino</button>

      <h2 style={{ marginTop: 20 }}>Seus Treinos</h2>
      {treinos.length === 0 ? (
        <p>Você ainda não tem treinos.</p>
      ) : (
        treinos.map(t => (
          <div
            key={t.id}
            style={{ border: '1px solid #ccc', padding: 10, marginTop: 10 }}
          >
            <strong>
              {t.divisao} – {t.nivel} / {t.objetivo}
            </strong>
            <br />
            Dias: {t.dias}
            <br />
            Exercícios: {t.exercicios.join(', ')}
          </div>
        ))
      )}
    </div>
  );
}
