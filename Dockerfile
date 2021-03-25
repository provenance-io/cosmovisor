# Create the cosmovisor source image
FROM golang:1.15-buster as build
WORKDIR /app
COPY ./go.* ./
RUN go mod download
COPY . .
RUN go install ./cmd/cosmovisor

# Final image here is ~30M
FROM gcr.io/distroless/base-debian10 as run
COPY --from=build /go/bin/cosmovisor /usr/bin/cosmovisor
ENTRYPOINT ["/usr/bin/cosmovisor"]
