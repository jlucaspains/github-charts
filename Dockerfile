
FROM golang:1.22.4-alpine3.20 AS builder
WORKDIR /app

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

ENV USER=appuser
ENV UID=10001 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"
    
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w" -o ./github-charts ./main.go


FROM node:22.2.0-alpine3.20 AS svelteBuiler
WORKDIR /app
COPY frontend/ ./
RUN npm install --ignore-scripts
RUN echo "VITE_API_BASE_PATH=/api" > .env
RUN npm run build

FROM scratch as runner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
WORKDIR /app
COPY --from=builder /app/github-charts .
COPY --from=svelteBuiler /app/dist/ ./public/
USER appuser:appuser
EXPOSE 8000
ENTRYPOINT ["./github-charts"]