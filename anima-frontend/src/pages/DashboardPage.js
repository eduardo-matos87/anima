// Arquivo: anima-frontend/src/pages/DashboardPage.js

import React, { useEffect, useState } from 'react';
import api from '../api';
import { useHistory } from 'react-router-dom';

/**
 * Página principal (dashboard) exibindo e criando treinos.
 */
export default function DashboardPage() {
  const [treinos, setTreinos] = useState([]);
  const history = useHistory();

  // Ao montar, tenta buscar os treinos; se falhar, volta ao login.
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

  // Chama o endpoint de criação e recarrega a lista.
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

  // Remove o JWT e retorna ao login
  const handleLogout = () => {
    localStorage.removeItem('jwt');
    history.push('/login');
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Painel de Treinos</h1>
      {/* Botão de logout usando handleLogout */}
      <button onClick={handleLogout} style={{ float: 'right' }}>
        Sair
      </button>

      <button onClick={createTreino} style={{ marginTop: 20 }}>
        Criar Treino
      </button>

      <h2 style={{ marginTop: 40 }}>Seus Treinos</h2>
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
