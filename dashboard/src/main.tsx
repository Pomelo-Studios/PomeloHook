import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import App from './App.tsx'
import { AdminApp } from './AdminApp.tsx'
import { OrgApp } from './OrgApp.tsx'
import { ToastProvider } from './components/ui'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ToastProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<App />} />
          <Route path="/admin/*" element={<AdminApp />} />
          <Route path="/app/*" element={<OrgApp />} />
        </Routes>
      </BrowserRouter>
    </ToastProvider>
  </StrictMode>,
)
