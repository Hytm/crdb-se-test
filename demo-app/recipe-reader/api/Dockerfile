FROM golang:1.17 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o recipe-reader .

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/recipe-reader /recipe-reader

CMD ["/recipe-reader"]

# docker run -d -p 8091:8091 -e GIN_MODE=release -e DATABASE_URL="postgres://hungryroach@haproxy:26257/recipes?sslmode=disable" --net roachnet --link haproxy --name recipe-reader recipe-reader