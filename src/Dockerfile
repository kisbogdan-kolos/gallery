# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.26 AS build-stage

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /gallery

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /gallery /gallery

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/gallery"]