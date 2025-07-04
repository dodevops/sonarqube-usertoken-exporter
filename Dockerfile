FROM golang:1.24 as Builder

WORKDIR /compile
COPY . .
RUN go build -o sonarqube-usertoken-exporter cmd/main.go

FROM alpine:latest

RUN apk add gcompat
RUN adduser -D -u 1000 user

WORKDIR /app
COPY --from=Builder /compile/sonarqube-usertoken-exporter ./
RUN chmod +x sonarqube-usertoken-exporter

USER 1000

ENTRYPOINT exec sonarqube-usertoken-exporter