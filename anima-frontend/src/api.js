// Arquivo: anima-frontend/src/api.js

import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080',      // Sua API Go está rodando aqui
  headers: { 'Content-Type': 'application/json' }
});

// Interceptor para anexar o JWT em todas as requisições
api.interceptors.request.use(config => {
  const token = localStorage.getItem('jwt');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
