import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      port: (() => {
        const parsedPort = parseInt(env.VITE_PORT, 10);
        return Number.isNaN(parsedPort) ? 5173 : parsedPort;
	    })()
    },
    base: env.VITE_BASE_PATH || '/'
  }
})
