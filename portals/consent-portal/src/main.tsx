import {StrictMode} from 'react'
import {createRoot} from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import {AsgardeoProvider} from "@asgardeo/react";
import {BrowserRouter} from 'react-router-dom';
import {ConsentProvider} from "./ConsentContext.tsx";

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

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AsgardeoProvider
      baseUrl={window?.configs?.VITE_BASE_URL}
      clientId={window?.configs?.VITE_CLIENT_ID}
      afterSignInUrl={window?.configs?.signInRedirectURL}
      afterSignOutUrl={window?.configs?.signOutRedirectURL}
      scopes={window?.configs?.VITE_SCOPE}
      organizationHandle={"carbon.super"}
    >
      <BrowserRouter>
        <ConsentProvider>
          <App/>
        </ConsentProvider>
      </BrowserRouter>
    </AsgardeoProvider>
  </StrictMode>,
)
