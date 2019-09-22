# Create builder image
FROM golang:alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -a -o main cmd/gateway/main.go

# Create run image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/main .
EXPOSE 8080
EXPOSE 9090
CMD ["./main"] 
