// Arquivo: anima-frontend/src/api.js

import axios from 'axios';

const api = axios.create({
  // Troque localhost por seu IP na rede local
  baseURL: 'http://192.168.173.249:8080',
  headers: { 'Content-Type': 'application/json' }
});

api.interceptors.request.use(config => {
  const token = localStorage.getItem('jwt');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export default api;
