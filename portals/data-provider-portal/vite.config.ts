import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3001, // Frontend dev server port
    proxy: {
      // Proxy requests from /api to the backend server
      '/api': {
        target: 'http://localhost:3000', // The address of api-server
        changeOrigin: true,
        // Rewrite the path: remove the /api prefix before forwarding
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
