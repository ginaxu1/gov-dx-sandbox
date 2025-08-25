/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CONSENT_ENGINE_PATH: string;
  readonly VITE_BASE_PATH: string;
  readonly VITE_PORT: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}