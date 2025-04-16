// Arquivo: anima-frontend/src/pages/LogoutPage.js

import { useEffect } from 'react';
import { useHistory } from 'react-router-dom';

/**
 * Limpa o token e redireciona para /login assim que o componente monta.
 */
export default function LogoutPage() {
  const history = useHistory();

  useEffect(() => {
    localStorage.removeItem('jwt');
    history.replace('/login');
  }, [history]);

  return null; // não renderiza nada
}
