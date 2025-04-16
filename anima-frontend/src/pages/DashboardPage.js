// Arquivo: src/pages/DashboardPage.js
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';

export default function DashboardPage() {
  const [treinos, setTreinos] = useState([]);
  const navigate = useNavigate();
  const token = localStorage.getItem('jwt');

  useEffect(() => {
    if (!token) return navigate('/login');
    fetchTreinos();
  }, [token, navigate]);

  const fetchTreinos = async () => {
    try {
      const resp = await api.get('/treinos');
      setTreinos(resp.data);
    } catch {
      navigate('/login');
    }
  };

  const createTreino = async () => {
    try {
      await api.post('/treino/criar', {
        nivel: 'iniciante',
        objetivo: 'emagrecimento',
        dias: 3,
        divisao: 'A',
        exercicios: [1,2,11]
      });
      fetchTreinos();
    } catch {
      alert('Erro ao criar treino');
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Painel de Treinos</h1>
      <button onClick={createTreino}>Criar Treino</button>
      <h2 style={{ marginTop: 20 }}>Seus Treinos</h2>
      {treinos.length === 0 && <p>Sem treinos</p>}
      {treinos.map(t => (
        <div key={t.id} style={{border:'1px solid #ccc', padding:10, marginTop:10}}>
          <strong>{t.divisao} – {t.nivel}/{t.objetivo}</strong><br/>
          Dias: {t.dias}<br/>
          Exercícios: {t.exercicios.join(', ')}
        </div>
      ))}
    </div>
  );
}
