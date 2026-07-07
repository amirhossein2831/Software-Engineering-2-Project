FROM golang:1.26 AS build
WORKDIR /src
ARG GOPROXY=https://goproxy.cn
ENV GOPROXY=${GOPROXY} GOSUMDB=off
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/app /usr/local/bin/app
USER app
EXPOSE 8084
ENTRYPOINT ["app"]
