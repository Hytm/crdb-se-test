FROM golang:1.17 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o feed-manager .

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/feed-manager /feed-manager

CMD ["/feed-manager"]

# docker run -d -p 8090:8090 -e GIN_MODE=release -e DATABASE_URL="postgres://hungryroach@haproxy:26257/recipes?sslmode=disable" --net roachnet --link haproxy --name feed-manager feed-manager