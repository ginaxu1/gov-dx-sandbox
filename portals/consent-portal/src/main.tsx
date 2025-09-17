import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { AuthProvider } from "@asgardeo/auth-react";

declare global {
  interface Window {
    configs?: {
      apiUrl: string;
      VITE_CLIENT_ID: string;
      VITE_BASE_URL: string;
      VITE_SCOPE: string;
    };
  }
}
const config = {
     signInRedirectURL: "http://localhost:5173",
     signOutRedirectURL: "http://localhost:5173",
     clientID: window?.configs?.VITE_CLIENT_ID || import.meta.env.VITE_CLIENT_ID,
     baseUrl: window?.configs?.VITE_BASE_URL || import.meta.env.VITE_BASE_URL,
     scope: window?.configs?.VITE_SCOPE ? window.configs.VITE_SCOPE.split(',') : import.meta.env.VITE_SCOPE?.split(',') || ['openid', 'profile']
};
console.log("config", config);
console.log("window.configs", window.configs);
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </StrictMode>,
)
