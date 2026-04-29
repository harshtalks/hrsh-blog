FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hrsh-ssh .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

ENV TERM=xterm-256color
ENV COLORTERM=truecolor

WORKDIR /app
COPY --from=builder /app/hrsh-ssh .

RUN mkdir -p .ssh

EXPOSE 2222

CMD ["./hrsh-ssh"]
