# Frontend Dockerfile for Dante GPU Web App (Next.js)

# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Install dependencies based on the lock file
COPY package.json /app/
COPY package-lock.json /app/
RUN npm ci

# Copy the rest of the application code
COPY . .

# Build the Next.js application
RUN npm run build

# Stage 2: Production image
FROM node:18-alpine

WORKDIR /app

# Create non-root user
RUN adduser -S nextjs -u 1001

# Copy built assets from the builder stage
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
# COPY --from=builder /app/next.config.js ./ # If you have a next.config.js

# Install only production dependencies
RUN npm ci --only=production && npm cache clean --force

# Change ownership to nextjs user
RUN chown -R nextjs:nextjs /app
USER nextjs

EXPOSE 3000

ENV PORT 3000
# ENV NEXT_TELEMETRY_DISABLED 1 # Optional: Disable Next.js telemetry

CMD ["npm", "start"] # Assumes your package.json has a "start": "next start" script 