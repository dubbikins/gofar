# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY . .
RUN go mod download
RUN go build -o /gofar

EXPOSE 8080

CMD [ "/gofar" ]
