// anima-frontend/src/pages/DashboardPage.js

import React, { useState, useEffect } from 'react';
import api from '../api';
import { Plus, Loader2 } from 'lucide-react';

export default function DashboardPage() {
  const [objetivos, setObjetivos] = useState([]);
  const [objetivoSelecionado, setObjetivoSelecionado] = useState('');
  const [treino, setTreino] = useState(null);
  const [loadingObjetivos, setLoadingObjetivos] = useState(false);
  const [loadingTreino, setLoadingTreino] = useState(false);

  useEffect(() => {
    setLoadingObjetivos(true);
    api.get('/objetivos')
      .then(res => setObjetivos(res.data))
      .catch(() => {})
      .finally(() => setLoadingObjetivos(false));
  }, []);

  async function handleSugerirTreino() {
    if (!objetivoSelecionado) return;
    setLoadingTreino(true);
    try {
      const { data } = await api.post('/gerar-treino', { objetivo: objetivoSelecionado });
      setTreino(data);
    } catch {
      setTreino(null);
    } finally {
      setLoadingTreino(false);
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 to-white py-12 px-4">
      <div className="max-w-3xl mx-auto space-y-8">

        {/* HEADER */}
        <header className="text-center">
          <h1 className="text-4xl font-extrabold text-indigo-800">Painel de Treinos</h1>
        </header>

        {/* AÇÃO: Criar Treino */}
        <div className="flex justify-end">
          <button
            className="inline-flex items-center space-x-2 bg-indigo-600 hover:bg-indigo-700 text-white px-5 py-2 rounded-full shadow-lg transition"
            onClick={() => { /* abrir modal ou rota */ }}
          >
            <Plus className="w-5 h-5" />
            <span>Criar Treino</span>
          </button>
        </div>

        {/* SELETOR DE OBJETIVO */}
        <div className="bg-white rounded-2xl shadow-xl p-6">
          <h2 className="text-xl font-semibold text-gray-700 mb-4">
            Selecione um Objetivo
          </h2>

          <div className="flex gap-4 flex-col sm:flex-row sm:items-center">
            <select
              className="flex-1 border border-gray-300 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-300"
              value={objetivoSelecionado}
              onChange={e => setObjetivoSelecionado(e.target.value)}
              disabled={loadingObjetivos}
            >
              <option value="">-- escolha um objetivo --</option>
              {objetivos.map(o => (
                <option key={o.id} value={o.nome}>{o.nome}</option>
              ))}
            </select>

            <button
              className={`inline-flex items-center space-x-2 px-4 py-2 rounded-lg shadow 
                ${objetivoSelecionado 
                  ? 'bg-green-500 hover:bg-green-600 text-white' 
                  : 'bg-gray-200 text-gray-500 cursor-not-allowed'}`}
              onClick={handleSugerirTreino}
              disabled={!objetivoSelecionado || loadingTreino}
            >
              {loadingTreino
                ? <Loader2 className="animate-spin w-5 h-5" />
                : <span>Sugerir Treinos</span>
              }
            </button>
          </div>
        </div>

        {/* RESULTADO: Seus Treinos */}
        <div className="bg-white rounded-2xl shadow-xl p-6">
          <h2 className="text-xl font-semibold text-gray-700 mb-4">Seus Treinos</h2>
          
          {!treino && (
            <p className="text-gray-500">Você ainda não tem treinos.</p>
          )}

          {treino && (
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="font-medium">Objetivo:</span>
                <span className="text-gray-600">{treino.objetivo}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="font-medium">Nível:</span>
                <span className="text-gray-600">{treino.nivel || '–'}</span>
              </div>

              <ul className="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-2">
                {treino.exercicios.map((ex, i) => (
                  <li
                    key={i}
                    className="flex items-center space-x-2 bg-indigo-50 rounded-lg p-3"
                  >
                    <span className="w-2 h-2 bg-indigo-600 rounded-full inline-block" />
                    <span className="text-gray-800">{ex}/span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
