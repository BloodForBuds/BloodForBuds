FROM golang:1.27-alpine AS base

WORKDIR /src

COPY go.mod ./
RUN go mod download

FROM base AS development

RUN go install github.com/air-verse/air@v1.67.3

COPY . .

CMD ["air", "-c", ".air.toml"]

FROM base AS build

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

FROM alpine:3.24 AS production

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /out/server /usr/local/bin/server

USER app

EXPOSE 8080

ENTRYPOINT ["server"]
