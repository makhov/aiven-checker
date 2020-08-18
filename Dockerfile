FROM golang:1.14 AS build

WORKDIR /app
COPY . /app

RUN go build -mod=vendor -o /checker ./cmd/checker
RUN go build -mod=vendor -o /writer ./cmd/writer

FROM debian:jessie-slim

COPY --from=build /checker /checker
COPY --from=build /writer /writer
COPY ./migrations /migrations

