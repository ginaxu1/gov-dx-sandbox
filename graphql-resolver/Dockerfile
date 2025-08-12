# Use official Node.js LTS version as base
FROM node:20-alpine

# Set working directory
WORKDIR /usr/src/app

# Copy package.json and package-lock.json first (better caching)
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy application source code
COPY . .

# Expose app port (change if your app uses a different port)
EXPOSE 4000

# Start the application
CMD ["npm", "start"]