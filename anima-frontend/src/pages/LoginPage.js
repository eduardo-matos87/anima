// Arquivo: anima-frontend/src/pages/LoginPage.js

import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';  // v5 usa useHistory()
import api from '../api';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const history = useHistory();  // substitui useNavigate()

  const handleSubmit = async e => {
    e.preventDefault();
    try {
      const resp = await api.post('/login', { email, password });
      localStorage.setItem('jwt', resp.data.token);
      history.push('/dashboard');  // redireciona para /dashboard
    } catch {
      setError('Falha no login. Verifique suas credenciais.');
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="email" placeholder="Email"
          value={email} onChange={e => setEmail(e.target.value)} required
        />
        <br /><br />
        <input
          type="password" placeholder="Senha"
          value={password} onChange={e => setPassword(e.target.value)} required
        />
        <br /><br />
        <button type="submit">Entrar</button>
      </form>
      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  );
}
