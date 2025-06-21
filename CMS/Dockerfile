# Dockerfile
FROM golang:1.23.4-alpine

WORKDIR /app

# Instala git (necesario para dependencias go get)
RUN apk add --no-cache git

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Por defecto compila la app (esto es opcional si sólo estás testeando)
RUN go build -o main .

# Permitir que podamos sobreescribir el CMD en docker-compose
CMD ["./main"]
