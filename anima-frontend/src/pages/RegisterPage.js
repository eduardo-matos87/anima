// Arquivo: anima-frontend/src/pages/RegisterPage.js

import React, { useState } from 'react';
import api from '../api';

export default function RegisterPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [msg, setMsg] = useState('');

  const handleSubmit = async e => {
    e.preventDefault();
    try {
      const resp = await api.post('/register', { name, email, password });
      setMsg(`Cadastrado com sucesso! ID: ${resp.data.user_id}`);
    } catch {
      setMsg('Falha no registro. Tente novamente.');
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>Registrar</h2>
      <form onSubmit={handleSubmit}>
        <input
          placeholder="Nome"
          value={name}
          onChange={e => setName(e.target.value)}
          required
        />
        <br /><br />
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
        <button type="submit">Registrar</button>
      </form>
      {msg && <p>{msg}</p>}
    </div>
  );
}
