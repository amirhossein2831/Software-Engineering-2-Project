FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/auth /usr/local/bin/auth
USER app
EXPOSE 8081
ENTRYPOINT ["auth"]
