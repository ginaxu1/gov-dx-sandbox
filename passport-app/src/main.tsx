import React from 'react';
import ReactDOM from 'react-dom/client';
import { AuthProvider } from '@asgardeo/auth-react';
import App from './App.tsx';
import './styles.css';

const config = {
  signInRedirectURL: window.location.origin,
  signOutRedirectURL: window.location.origin,
  clientID: import.meta.env.VITE_ASGARDEO_CLIENT_ID,
  baseUrl: import.meta.env.VITE_ASGARDEO_BASE_URL,
  scope: [import.meta.env.VITE_ASGARDEO_SCOPE],
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </React.StrictMode>
);