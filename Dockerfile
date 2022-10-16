FROM golang:1.19-buster AS build

WORKDIR /app

COPY . .

RUN go build -o /bin/app ./cmd/app/main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /bin/app /app
ENV GIN_MODE=release
EXPOSE 8080

ENTRYPOINT ["/app"]
