// src/api.js
import axios from 'axios';

// Cria uma instância do Axios com a URL base da sua API
const api = axios.create({
  baseURL: 'http://localhost:8080', // altere se a API estiver em outro endereço
});

// Adiciona um interceptor para incluir o token, se existir
api.interceptors.request.use((config) => {
  // Obtém o token do localStorage (ou de onde você armazenar)
  const token = localStorage.getItem('jwt');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});

export default api;
