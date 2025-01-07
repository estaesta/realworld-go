# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN make build

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN make test

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/main /

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/main"]
