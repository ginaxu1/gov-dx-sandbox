// main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { AuthProvider } from '@asgardeo/auth-react';
import App from './App.tsx';
import './styles.css';

const config = {
  signInRedirectURL: window.location.origin,
  signOutRedirectURL: window.location.origin,
  clientID: "gbzkiYFzfNRCYIK3EqEln3ntXM8a",
  baseUrl: "https://api.asgardeo.io/t/lankasoftwarefoundation",
  scope: ["openid", "profile"]
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <script src="/public/config.js"></script>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </React.StrictMode>
);