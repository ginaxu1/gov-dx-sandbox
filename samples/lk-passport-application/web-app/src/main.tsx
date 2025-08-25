// main.tsx
import React from "react";
import ReactDOM from "react-dom/client";
import { AuthProvider } from "@asgardeo/auth-react";
import App from "./App.tsx";
import "./styles.css";

// Get the base URL from the environment variable if it exists, otherwise use the local origin
const appBaseUrl = import.meta.env.VITE_APP_URL
  ? import.meta.env.VITE_APP_URL
  : window.location.origin;

const config = {
  signInRedirectURL: appBaseUrl,
  signOutRedirectURL: appBaseUrl,
  clientID: import.meta.env.VITE_ASGARDEO_CLIENT_ID,
  baseUrl: import.meta.env.VITE_ASGARDEO_BASE_URL,
  scope: import.meta.env.VITE_ASGARDEO_SCOPE?.split(" "),
  endpoints: {
    authorizeEndpoint: `${import.meta.env.VITE_ASGARDEO_BASE_URL}/oauth2/authorize`,
    tokenEndpoint: `${import.meta.env.VITE_ASGARDEO_BASE_URL}/oauth2/token`,
    jwksUri: `${import.meta.env.VITE_ASGARDEO_BASE_URL}/oauth2/jwks`,
  },
};

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <AuthProvider config={config}>
      <App />
    </AuthProvider>
  </React.StrictMode>,
);
