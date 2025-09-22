import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { AuthProvider } from "@asgardeo/auth-react";

declare global {
  interface Window {
    configs: {
      apiUrl: string;
      VITE_CLIENT_ID: string;
      VITE_BASE_URL: string;
      VITE_SCOPE: string;
      signInRedirectURL: string;
      signOutRedirectURL: string;
    };
  }
}
const config = {
     signInRedirectURL: window?.configs?.signInRedirectURL || "http://localhost:5173",
     signOutRedirectURL: window?.configs?.signOutRedirectURL || "http://localhost:5173",
     clientID: window?.configs?.VITE_CLIENT_ID || import.meta.env.VITE_ASGARDEO_CLIENT_ID,
     baseUrl: window?.configs?.VITE_BASE_URL || import.meta.env.VITE_ASGARDEO_BASE_URL,
     scope: window?.configs?.VITE_SCOPE ? window.configs.VITE_SCOPE.split(',') : import.meta.env.VITE_ASGARDEO_SCOPE?.split(',') || ['openid', 'profile'],
     endpoints: {
         authorizationEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/authorize",
         tokenEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token",
         userInfoEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/userinfo",
         endSessionEndpoint: "https://api.asgardeo.io/t/lankasoftwarefoundation/oidc/logout"
     }
};
console.log("Auth config:", config);
console.log("Environment variables:", {
  VITE_ASGARDEO_CLIENT_ID: import.meta.env.VITE_ASGARDEO_CLIENT_ID,
  VITE_ASGARDEO_BASE_URL: import.meta.env.VITE_ASGARDEO_BASE_URL,
  VITE_ASGARDEO_SCOPE: import.meta.env.VITE_ASGARDEO_SCOPE
});
console.log("Window configs:", window.configs);
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </StrictMode>,
)
