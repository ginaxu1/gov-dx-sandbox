// main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { AuthProvider } from '@asgardeo/auth-react';
import App from './App.tsx';
import './styles.css';

const config = {
  signInRedirectURL: "https://7a6923cc-c4a5-44eb-bdd7-90d84414cb83.e1-us-east-azure.choreoapps.dev/",
  signOutRedirectURL: "https://7a6923cc-c4a5-44eb-bdd7-90d84414cb83.e1-us-east-azure.choreoapps.dev/",
  clientID: "gbzkiYFzfNRCYIK3EqEln3ntXM8a",
  baseUrl: "https://api.asgardeo.io/t/lankasoftwarefoundation",
  scope: ["openid", "profile"],
  authorizeEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/authorize",
  tokenEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token",
  jwksUri: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks",
};  

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </React.StrictMode>
);