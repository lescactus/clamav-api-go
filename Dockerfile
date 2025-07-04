FROM golang:1.24 as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags '-d -w -s' -o main

FROM scratch

COPY --from=builder /app/main /

EXPOSE 8080

CMD ["/main"]