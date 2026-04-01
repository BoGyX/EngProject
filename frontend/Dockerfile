FROM node:20-alpine AS dev

WORKDIR /app

COPY package*.json ./
RUN npm ci

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

FROM node:20-alpine AS builder

WORKDIR /app

# Копируем package-файлы и ставим зависимости до копирования исходников
COPY package*.json ./
RUN npm ci

# Копируем исходники фронтенда и собираем production bundle
COPY . .
ARG VITE_API_URL=/api
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=80
ENV VITE_API_URL=/api
ENV BACKEND_ORIGIN=http://backend:8000
ENV FORCE_HTTPS_REDIRECT=false

COPY package.json ./package.json
COPY server.js ./server.js
COPY --from=builder /app/dist ./dist

EXPOSE 80

CMD ["node", "/app/server.js"]
