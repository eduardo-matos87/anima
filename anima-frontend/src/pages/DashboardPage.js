// anima-frontend/src/pages/DashboardPage.js

import React, { useState, useEffect } from 'react';
import api from '../api'; // seu axios configurado com baseURL e token

export default function DashboardPage() {
  const [objetivos, setObjetivos] = useState([]);
  const [objetivoSelecionado, setObjetivoSelecionado] = useState('');
  const [treino, setTreino] = useState(null);
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState('');

  // 1. Carrega objetivos ao montar o componente
  useEffect(() => {
    async function fetchObjetivos() {
      try {
        const { data } = await api.get('/objetivos');
        setObjetivos(data);
      } catch (e) {
        console.error(e);
      }
    }
    fetchObjetivos();
  }, []);

  // 2. Trata mudança no select
  function handleChangeObjetivo(e) {
    setObjetivoSelecionado(e.target.value);
    setTreino(null);
    setErro('');
  }

  // 3. Chama o back‑end para gerar o treino
  async function handleSugerirTreino() {
    if (!objetivoSelecionado) return;
    setCarregando(true);
    setErro('');
    try {
      const { data } = await api.post('/gerar-treino', {
        objetivo: objetivoSelecionado,
        // se precisar de nível, adicione aqui
      });
      setTreino(data);
    } catch (e) {
      console.error(e);
      setErro('Erro ao gerar treino');
    } finally {
      setCarregando(false);
    }
  }

  return (
    <div className="max-w-xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Painel de Treinos</h1>

      {/* Criar Treino (pode abrir modal ou navegar) */}
      <button
        className="mb-6 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        onClick={() => {/* navegue ou abra modal */}}
      >
        Criar Treino
      </button>

      {/* Seletor de Objetivo */}
      <div className="mb-4">
        <h2 className="text-lg font-semibold mb-2">
          Selecione um Objetivo para Sugestões
        </h2>
        <select
          className="border px-3 py-2 rounded w-full"
          value={objetivoSelecionado}
          onChange={handleChangeObjetivo}
        >
          <option value="">-- Escolha um objetivo --</option>
          {objetivos.map(obj => (
            <option key={obj.id} value={obj.nome}>
              {obj.nome}
            </option>
          ))}
        </select>
        <button
          className={`mt-2 px-4 py-2 rounded ${
            objetivoSelecionado
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : 'bg-gray-300 text-gray-600 cursor-not-allowed'
          }`}
          disabled={!objetivoSelecionado || carregando}
          onClick={handleSugerirTreino}
        >
          {carregando ? 'Carregando...' : 'Sugerir Treinos'}
        </button>
        {erro && <p className="text-red-500 mt-2">{erro}</p>}
      </div>

      {/* Exibe os treinos gerados */}
      <div>
        <h2 className="text-lg font-semibold mb-2">Seus Treinos</h2>
        {!treino && <p>Você ainda não tem treinos.</p>}
        {treino && (
          <div className="bg-gray-100 p-4 rounded">
            <p className="font-medium">
              Objetivo: <span className="font-normal">{treino.objetivo}</span>
            </p>
            <p className="font-medium">
              Nível: <span className="font-normal">{treino.nivel || '—'}</span>
            </p>
            <ul className="list-disc list-inside mt-2">
              {treino.exercicios.map((ex, i) => (
                <li key={i}>{ex}</li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
