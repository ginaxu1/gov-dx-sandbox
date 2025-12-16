import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { AuthProvider, type AuthProviderProps } from "react-oidc-context";
import { BrowserRouter } from 'react-router-dom';
import App from './App.tsx';
import { ConsentProvider } from './contexts/ConsentContext';
import './index.css';

declare global {
  interface Window {
    configs: {
      apiUrl: string;
      VITE_CLIENT_ID: string;
      VITE_BASE_URL: string;
      VITE_SCOPE: string;
      signInRedirectURL: string;
      signOutRedirectURL: string;
      organizationHandle: string;
    };
  }
}

const oidcConfig: AuthProviderProps = {
  authority: window?.configs?.VITE_BASE_URL, // Used as base for validation, but metadata overrides endpoints
  client_id: window?.configs?.VITE_CLIENT_ID,
  redirect_uri: window?.configs?.signInRedirectURL,
  post_logout_redirect_uri: window?.configs?.signOutRedirectURL,
  scope: window?.configs?.VITE_SCOPE || 'openid profile email',
  // Manually define endpoints to match previous Asgardeo config exactly
  metadata: {
    authorization_endpoint: `${window?.configs?.VITE_BASE_URL}/oauth2/authorize`,
    token_endpoint: `${window?.configs?.VITE_BASE_URL}/oauth2/token`,
    end_session_endpoint: `${window?.configs?.VITE_BASE_URL}/oidc/logout`,
    jwks_uri: `${window?.configs?.VITE_BASE_URL}/oauth2/jwks`,
    revocation_endpoint: `${window?.configs?.VITE_BASE_URL}/oauth2/revoke`,
    check_session_iframe: `${window?.configs?.VITE_BASE_URL}/oidc/checksession`,
    userinfo_endpoint: `${window?.configs?.VITE_BASE_URL}/oauth2/userinfo`,
    issuer: `${window?.configs?.VITE_BASE_URL}/oauth2/token`,
  },
  onSigninCallback: () => {
    // Remove query params (code, state) after successful login
    window.history.replaceState({}, document.title, window.location.pathname);
  }
};

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProvider {...oidcConfig}>
      <BrowserRouter>
        <ConsentProvider>
          <App />
        </ConsentProvider>
      </BrowserRouter>
    </AuthProvider>
  </StrictMode>,
)
