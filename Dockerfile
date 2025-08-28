FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код для сборки
COPY . . 

# Сборка приложения
RUN go build -o apiserver ./cmd/main.go

# --- Финальный образ ---
# ... начало вашего Dockerfile ...

 # --- Финальный образ ---
 FROM alpine:latest

 WORKDIR /app

 # Копирование скомпилированного бинарника
 COPY --from=builder /app/apiserver .

 # Копирование схемы всё ещё может быть нужно, если вы её используете
 COPY --from=builder /app/schema ./schema
 
 # ... остальная часть Dockerfile ...
 CMD ["./apiserver"]