/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CONSENT_ENGINE_PATH: string;
  readonly VITE_BASE_PATH: string;
  readonly VITE_PORT: string;
  readonly VITE_CLIENT_ID: string;
  readonly VITE_BASE_URL: string;
  readonly VITE_SCOPE: string;
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}