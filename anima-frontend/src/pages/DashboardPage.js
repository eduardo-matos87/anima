import React, { useEffect, useState } from 'react';
import api from '../api';
import { useHistory } from 'react-router-dom';

export default function DashboardPage() {
  const [treinos, setTreinos] = useState([]);
  const [objetivos, setObjetivos] = useState([]);
  const [selectedObjetivo, setSelectedObjetivo] = useState('');
  const [sugestoes, setSugestoes] = useState([]);
  const history = useHistory();

  // Busca objetivos e treinos ao montar
  useEffect(() => {
    const init = async () => {
      try {
        const [objResp, trResp] = await Promise.all([
          api.get('/objetivos'),
          api.get('/treinos')
        ]);
        setObjetivos(objResp.data);
        setTreinos(trResp.data);
      } catch {
        history.push('/login');
      }
    };
    init();
  }, [history]);

  // Lida com criação de treino e recarrega lista
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

  // Solicita sugestões baseado no objetivo selecionado
  const handleSugestao = async () => {
    try {
      const resp = await api.get(`/treino?objetivo=${encodeURIComponent(selectedObjetivo)}`);
      setSugestoes(resp.data);
    } catch {
      alert('Erro ao buscar sugestões');
    }
  };

  // Logout
  const handleLogout = () => {
    localStorage.removeItem('jwt');
    history.push('/login');
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Painel de Treinos</h1>
      <button onClick={handleLogout} style={{ float: 'right' }}>Sair</button>

      <section style={{ marginBottom: 40 }}>
        <h2>Criar Treino</h2>
        <button onClick={createTreino}>Criar Treino</button>
      </section>

      <section style={{ marginBottom: 40 }}>
        <h2>Selecione um Objetivo para Sugestões</h2>
        <select
          value={selectedObjetivo}
          onChange={e => setSelectedObjetivo(e.target.value)}
          style={{ padding: 8, marginRight: 10 }}
        >
          <option value="">-- Escolha um objetivo --</option>
          {objetivos.map(o => (
            <option key={o.id} value={o.nome}>{o.nome}</option>
          ))}
        </select>
        <button
          type="button"
          disabled={!selectedObjetivo}
          onClick={handleSugestao}
          style={{ padding: '8px 16px' }}
        >
          Sugerir Treinos
        </button>

        {sugestoes.length > 0 && (
          <div style={{ marginTop: 20 }}>
            <h3>Sugestões para {selectedObjetivo}</h3>
            {sugestoes.map(t => (
              <div
                key={t.id}
                style={{ border: '1px solid #8b8', padding: 10, margin: '10px 0' }}
              >
                <strong>{t.divisao} – {t.nivel}</strong><br />
                Dias: {t.dias}<br />
                Exercícios: {t.exercicios.join(', ')}
              </div>
            ))}
          </div>
        )}
      </section>

      <section>
        <h2>Seus Treinos</h2>
        {treinos.length === 0 ? (
          <p>Você ainda não tem treinos.</p>
        ) : (
          treinos.map(t => (
            <div
              key={t.id}
              style={{ border: '1px solid #ccc', padding: 10, marginTop: 10 }}
            >
              <strong>{t.divisao} – {t.nivel} / {t.objetivo}</strong><br />
              Dias: {t.dias}<br />
              Exercícios: {t.exercicios.join(', ')}
            </div>
          ))
        )}
      </section>
    </div>
  );
}
