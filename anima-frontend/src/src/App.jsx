import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";

// imports que você já tem:
import LoginPage from "./pages/LoginPage";           // se existir
import HomePage from "./pages/HomePage";             // se existir

// nosso gerador
import DashboardPage from "./pages/DashboardPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* suas rotas existentes */}
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />

        {/* ✅ atalho público para o gerador de treinos */}
        <Route path="/treinos" element={<DashboardPage />} />

        {/* 404 opcional */}
        <Route path="*" element={<div style={{padding:24}}>Página não encontrada</div>} />
      </Routes>
    </BrowserRouter>
  );
}
