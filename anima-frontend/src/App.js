// Arquivo: src/App.js

import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* redireciona quem já estiver logado */}
        <Route
          path="/login"
          element={
            localStorage.getItem('jwt') ? <Navigate to="/dashboard"/> : <LoginPage/>
          }
        />
        <Route
          path="/register"
          element={
            localStorage.getItem('jwt') ? <Navigate to="/dashboard"/> : <RegisterPage/>
          }
        />
        <Route
          path="/dashboard"
          element={<DashboardPage/>}
        />
        {/* qualquer outra rota vai para /login */}
        <Route path="*" element={<Navigate to="/login"/>} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
