FROM golang:1.27-bookworm AS builder
  
WORKDIR /app
  
COPY go.mod go.sum ./
RUN go mod download
  
COPY . .
  
RUN CGO_ENABLED=0 GOOS=linux go build -o getcourse-schedulix ./cmd/bot
  

FROM debian:bookworm-slim
  
RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		libreoffice \
		poppler-utils \
	&& rm -rf /var/lib/apt/lists/*
  
WORKDIR /app
  
COPY --from=builder /app/getcourse-schedulix .
  
CMD ["./getcourse-schedulix"]