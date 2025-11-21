# Build Stage
FROM node:20-alpine AS build

WORKDIR /app

# Declare build arguments for environment variables
# These are needed at build time for Vite to embed them in the bundle
ARG VITE_PORT=5173
ARG VITE_BASE_PATH=/

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

# Serve Stage
FROM nginx:1.27-alpine

# Copy built static files from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set default environment variables (can be overridden at runtime)
ENV VITE_API_URL=""
ENV VITE_CLIENT_ID=""
ENV VITE_BASE_URL=""
ENV VITE_SCOPE=""
ENV VITE_SIGN_IN_REDIRECT_URL=""
ENV VITE_SIGN_OUT_REDIRECT_URL=""

# Expose port 80
EXPOSE 80

# Use entrypoint script to generate config.js at runtime
ENTRYPOINT ["/entrypoint.sh"]

