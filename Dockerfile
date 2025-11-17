# Build stage
FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache git gcc musl-dev tzdata

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG BUILD_TARGET=bot

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/${BUILD_TARGET} ./cmd/${BUILD_TARGET}

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

ENV TZ=Europe/Moscow

ARG BUILD_TARGET=bot

COPY --from=builder /app/bin/${BUILD_TARGET} ./${BUILD_TARGET}

RUN mkdir -p /root/data

EXPOSE 2000

CMD ./${BUILD_TARGET}
