FROM golang:1.23.0-alpine3.20 AS builder
RUN apk --no-cache add tzdata ca-certificates git

WORKDIR /src/
COPY . /src/

RUN go build -a -installsuffix cgo -o /app ./main.go

FROM alpine:3.20.3
COPY --from=builder /app /app
COPY entrypoint.sh /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]