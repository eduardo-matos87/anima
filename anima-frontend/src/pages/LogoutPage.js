// Arquivo: anima-frontend/src/pages/LogoutPage.js

import React, { useEffect } from 'react';
import { useHistory } from 'react-router-dom';

export default function LogoutPage() {
  const history = useHistory();

  useEffect(() => {
    // Remove o token
    localStorage.removeItem('jwt');
    // Redireciona ao login
    history.replace('/login');
  }, [history]);

  return null; // não precisa renderizar nada
}
