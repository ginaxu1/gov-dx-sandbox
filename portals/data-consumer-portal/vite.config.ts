import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  // Load environment variables from .env files based on the current mode (development, production)
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [
      react(), 
    ],
    server: {
      // Set the port by reading the VITE_PORT variable from the .env file.
      // If it's not defined or invalid, it will default to 5173.
      port: (() => {
        const parsedPort = parseInt(env.VITE_PORT, 10);
        return Number.isNaN(parsedPort) ? 5173 : parsedPort;
      })(),
      
      // Keep the proxy configuration to forward API requests to the backend server
      proxy: {
        '/api': {
          target: 'http://localhost:3000', // backend server address
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
        },
      },
    },
    // Set the base path for the application, also read from the .env file
    base: env.VITE_BASE_PATH || '/'
  };
});

