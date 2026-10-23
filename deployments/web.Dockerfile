# Build frontend + publish with nginx, proxying /api to the backend
# Build context must be the REPO ROOT:
#   docker build -f deployments/web.Dockerfile .
FROM node:20-alpine AS builder
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM nginx:1.27-alpine
COPY deployments/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
