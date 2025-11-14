# ===== 1) Builder =====
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Опционально: укажи прокси/приватные модули при необходимости
ENV CGO_ENABLED=0 \
    GO111MODULE=on

# Сначала зависимости (лучший кэш)
COPY go.mod go.sum ./
RUN go mod download

# Затем исходники
COPY . .

# Создаем директорию для выходного файла
RUN mkdir -p /out

# Собираем статически с минификацией символов
# Указываем конкретный main пакет вместо ./...
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/app .

# ===== 2) Runtime (минимальный) =====
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Кладём бинарь
COPY --from=builder --chown=nonroot:nonroot /out/app /app/app
COPY --from=builder --chown=nonroot:nonroot /src/config /app/config
# Непривилегированный пользователь уже задан (nonroot)
USER nonroot:nonroot

# Приложение слушает порт 8080 (замени при необходимости)
EXPOSE 8080
ENTRYPOINT ["/app/app"]

