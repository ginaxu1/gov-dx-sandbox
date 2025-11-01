import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { AuthProvider } from "@asgardeo/auth-react";

declare global {
  interface Window {
    configs: {
      VITE_API_URL: string;
      VITE_LOGS_URL: string;
      VITE_IDP_CLIENT_ID: string;
      VITE_IDP_BASE_URL: string;
      VITE_IDP_SCOPE: string;
      VITE_IDP_ADMIN_ROLE: string;
      VITE_SIGN_IN_REDIRECT_URL: string;
      VITE_SIGN_OUT_REDIRECT_URL: string;
    };
  }
}

const config = {
     signInRedirectURL: window?.configs?.VITE_SIGN_IN_REDIRECT_URL,
     signOutRedirectURL: window?.configs?.VITE_SIGN_OUT_REDIRECT_URL,
     clientID: window?.configs?.VITE_IDP_CLIENT_ID,
     baseUrl: window?.configs?.VITE_IDP_BASE_URL,
     scope: window?.configs?.VITE_IDP_SCOPE ? window.configs.VITE_IDP_SCOPE.split(',') : ['openid', 'profile'],
     endpoints: {
         authorizationEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/authorize",
         tokenEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token",
         userInfoEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/userinfo",
         endSessionEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oidc/logout"
     }
};

console.log("Auth config:", config);
console.log("Window configs:", window.configs);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </StrictMode>,
)
