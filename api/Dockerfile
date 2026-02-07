FROM golang:1.25.4 AS build

WORKDIR /app

COPY . .

# Build a static binary
RUN go build -o api .

FROM debian:bookworm-slim

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=build /app/api .

CMD ["./api"]
