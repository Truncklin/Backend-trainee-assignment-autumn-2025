# build stage
FROM golang:1.25.1 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /src/cmd/pr-reviewer
RUN CGO_ENABLED=0 GOOS=linux go build -o /prservice .

# runtime
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /prservice /prservice
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/prservice"]
