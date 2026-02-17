# Multi-stage build for Next.js MQT Frontend
# Stage 1: Dependencies
FROM node:20-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

# Stage 2: Build
FROM node:20-alpine AS build
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build-time args for API URLs (baked into client bundle)
# Default to the production HTTPS API so that if build args
# are not overridden, the frontend still talks to the correct
# public endpoint instead of localhost.
ARG NEXT_PUBLIC_API_BASE_URL=https://orpheus-networks.com/api
ARG NEXT_PUBLIC_READINGS_API_BASE_URL=https://orpheus-networks.com/api
ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}
ENV NEXT_PUBLIC_READINGS_API_BASE_URL=${NEXT_PUBLIC_READINGS_API_BASE_URL}

RUN npm run build

# Stage 3: Runtime
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy standalone build output
COPY --from=build /app/public ./public
COPY --from=build --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
