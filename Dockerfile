FROM golang:1.19-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd
RUN go build -o /main .

FROM alpine:latest AS prod
WORKDIR /app
RUN apk update
COPY --from=build /main .
CMD ["./main"]