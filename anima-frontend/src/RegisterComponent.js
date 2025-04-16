// anima-frontend/src/RegisterComponent.js
import React, { useState } from 'react';
import api from './api';

export default function RegisterComponent() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [msg, setMsg] = useState('');

  const register = async () => {
    try {
      const res = await api.post('/register', { name, email, password });
      setMsg(`Usuário criado! ID: ${res.data.user_id}`);
    } catch (err) {
      setMsg('Falha no registro');
    }
  };

  return (
    <div>
      <h2>Cadastro</h2>
      <input placeholder="Nome" value={name} onChange={e=>setName(e.target.value)} />
      <input placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} />
      <input type="password" placeholder="Senha" value={password} onChange={e=>setPassword(e.target.value)} />
      <button onClick={register}>Registrar</button>
      {msg && <p>{msg}</p>}
    </div>
  );
}
