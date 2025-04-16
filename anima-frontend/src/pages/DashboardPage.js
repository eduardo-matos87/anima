// Arquivo: anima-frontend/src/pages/DashboardPage.js

import React, { useEffect, useState } from 'react';
import api from '../api';

export default function DashboardPage() {
  const [treinos, setTreinos] = useState([]);

  // Busca treinos ao montar o componente
  useEffect(() => {
    const fetchTreinos = async () => {
      try {
        const resp = await api.get('/treinos');
        setTreinos(resp.data);
      } catch (err) {
        console.error('Erro ao listar treinos:', err);
      }
    };
    fetchTreinos();
  }, []);

  const createTreino = async () => {
    try {
      await api.post('/treino/criar', {
        nivel: 'iniciante',
        objetivo: 'emagrecimento',
        dias: 3,
        divisao: 'A',
        exercicios: [1, 2, 11],
      });
      // recarrega lista depois de criar
      const resp = await api.get('/treinos');
      setTreinos(resp.data);
    } catch (err) {
      console.error('Erro ao criar treino:', err);
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Painel de Treinos</h1>
      <button onClick={createTreino}>Criar Treino</button>
      <h2 style={{ marginTop: 20 }}>Seus Treinos</h2>
      {treinos.length === 0 && <p>Você ainda não tem treinos.</p>}
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
