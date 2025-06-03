# Frontend Dockerfile for Dante GPU Web App (Next.js)

# Stage 1: Build the application
FROM node:18-alpine AS builder

WORKDIR /app

# Install pnpm globally if you use it, otherwise adjust for npm/yarn
# RUN npm install -g pnpm

# Copy package.json and lock file (pnpm-lock.yaml, package-lock.json, or yarn.lock)
COPY package.json /app/
# COPY pnpm-lock.yaml /app/ # Uncomment if using pnpm
# COPY package-lock.json /app/ # Uncomment if using npm and package-lock.json exists
# COPY yarn.lock /app/      # Uncomment if using yarn

# Install dependencies based on the lock file
# RUN npm ci

# Copy the rest of the application code
# COPY . .

# Build the Next.js application
# RUN npm run build

# Stage 2: Production image
FROM node:18-alpine

WORKDIR /app

# Create a non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

# Copy built assets from the builder stage
# COPY --from=builder /app/.next ./.next
# COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
# COPY --from=builder /app/next.config.js ./ # If you have a next.config.js
# COPY --from=builder /app/node_modules ./node_modules # Only if you have runtime deps not bundled by Next.js build

USER nextjs

EXPOSE 3000

ENV PORT 3000
# ENV NEXT_TELEMETRY_DISABLED 1 # Optional: Disable Next.js telemetry

CMD ["npm", "start"] # Assumes your package.json has a "start": "next start" script 