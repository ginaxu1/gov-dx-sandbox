# Build Stage
FROM node:20-alpine AS build

WORKDIR /app

# Declare build arguments for environment variables
# These are needed at build time for Vite to embed them in the bundle
ARG VITE_PORT=5173
ARG VITE_BASE_PATH=/

# Declare build arguments for config.js injection
# These will be used to generate config.js at build time
# Note: Consent portal expects config.js at ./public/config.js (different path)
ARG VITE_API_URL=""
ARG VITE_CLIENT_ID=""
ARG VITE_BASE_URL=""
ARG VITE_SCOPE=""
ARG VITE_SIGN_IN_REDIRECT_URL=""
ARG VITE_SIGN_OUT_REDIRECT_URL=""

# Set as environment variables for the build process
ENV VITE_PORT=${VITE_PORT}
ENV VITE_BASE_PATH=${VITE_BASE_PATH}

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application (Vite will use the environment variables)
RUN npm run build

# Generate config.js file from build arguments
# Note: Consent portal expects config.js at ./public/config.js (different from other portals)
# This file will be loaded by index.html and provide window.configs to the application
RUN mkdir -p /app/dist/public && \
    cat > /app/dist/public/config.js << EOF
window.configs = {
  apiUrl: '${VITE_API_URL}',
  VITE_CLIENT_ID: '${VITE_CLIENT_ID}',
  VITE_BASE_URL: '${VITE_BASE_URL}',
  VITE_SCOPE: '${VITE_SCOPE}',
  signInRedirectURL: '${VITE_SIGN_IN_REDIRECT_URL}',
  signOutRedirectURL: '${VITE_SIGN_OUT_REDIRECT_URL}'
};
EOF

# Serve Stage
FROM nginx:1.27-alpine

# Copy built static files from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

