FROM golang:1.21-alpine AS backend-builder

ARG VERSION=dev
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-linkmode external -extldflags '-static' -X github.com/bruno.lopes/calendar/backend/internal/api.Version=${VERSION}" -o server cmd/server/main.go

FROM node:20-alpine AS frontend-builder

ARG VERSION=dev
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ .
ENV VITE_APP_VERSION=${VERSION}
RUN npm run build

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata nginx supervisor

# Backend
WORKDIR /app
COPY --from=backend-builder /app/server .
RUN mkdir -p /app/data

# Frontend
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/http.d/default.conf

# Supervisor config to run both services
RUN mkdir -p /etc/supervisor.d
COPY <<EOF /etc/supervisor.d/services.ini
[supervisord]
nodaemon=true
logfile=/dev/stdout
logfile_maxbytes=0

[program:backend]
command=/app/server
directory=/app
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:nginx]
command=nginx -g "daemon off;"
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

EXPOSE 80

ENV GIN_MODE=release
ENV TZ=Europe/Lisbon

VOLUME ["/app/data"]

CMD ["supervisord", "-c", "/etc/supervisord.conf"]
