// Arquivo: anima-frontend/src/App.js

import React from 'react';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Redirect
} from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';

/**
 * Aqui usamos o Router da versão 5, com Switch/Route/Redirect.
 */
function App() {
  const isLogged = !!localStorage.getItem('jwt');

  return (
    <Router>
      <Switch>
        {/* Se já estiver logado, não deixa voltar ao login */}
        <Route path="/login">
          {isLogged ? <Redirect to="/dashboard" /> : <LoginPage />}
        </Route>

        <Route path="/register">
          {isLogged ? <Redirect to="/dashboard" /> : <RegisterPage />}
        </Route>

        <Route path="/dashboard">
          {isLogged ? <DashboardPage /> : <Redirect to="/login" />}
        </Route>

        {/* Qualquer outra URL redireciona para /login */}
        <Route path="*">
          <Redirect to="/login" />
        </Route>
      </Switch>
    </Router>
  );
}

export default App;
