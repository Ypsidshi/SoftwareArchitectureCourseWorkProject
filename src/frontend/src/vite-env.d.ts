/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_DEAL_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
