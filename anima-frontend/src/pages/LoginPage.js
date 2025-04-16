// Arquivo: anima-frontend/src/pages/LoginPage.js

import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';
import api from '../api';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const history = useHistory();

  const handleSubmit = async e => {
    e.preventDefault();

    // 🔍 Debug: URL e payload que serão enviados
    console.log('➡️ POST para:', api.defaults.baseURL + '/login');
    console.log('➡️ Payload:', { email, password });

    try {
      const resp = await api.post('/login', { email, password });
      console.log('✅ resposta do /login:', resp);

      // Salva token e redireciona para o dashboard
      localStorage.setItem('jwt', resp.data.token);
      history.push('/dashboard');
    } catch (err) {
      console.error('❌ erro no /login:', err.response || err);
      setError('Falha no login. Verifique suas credenciais.');
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={e => setEmail(e.target.value)}
          required
        />
        <br /><br />
        <input
          type="password"
          placeholder="Senha"
          value={password}
          onChange={e => setPassword(e.target.value)}
          required
        />
        <br /><br />
        <button type="submit">Entrar</button>
      </form>
      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  );
}
