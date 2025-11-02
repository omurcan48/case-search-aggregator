FROM golang:1.23 AS build
WORKDIR /src

# 1) Mod dosyalarını birlikte kopyala
COPY go.mod go.sum ./

# 2) Bağımlılıkları indir (cache için)
RUN go mod download

# 3) Kaynakları kopyala
COPY . .

# 4) Kaynaklar geldikten sonra mod dosyalarını hizala
RUN go mod tidy

# 5) Derle
RUN CGO_ENABLED=0 go build -o /app/api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/api /app/api
COPY web /app/web
COPY migrations /app/migrations
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/api"]