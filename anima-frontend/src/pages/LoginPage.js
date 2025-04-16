// Arquivo: src/pages/LoginPage.js
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async e => {
    e.preventDefault();
    try {
      const resp = await api.post('/login', { email, password });
      localStorage.setItem('jwt', resp.data.token);
      navigate('/dashboard');
    } catch {
      setError('Falha no login');
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="email" placeholder="Email"
          value={email} onChange={e => setEmail(e.target.value)} required
        /><br/><br/>
        <input
          type="password" placeholder="Senha"
          value={password} onChange={e => setPassword(e.target.value)} required
        /><br/><br/>
        <button type="submit">Entrar</button>
      </form>
      {error && <p style={{color:'red'}}>{error}</p>}
    </div>
  );
}
