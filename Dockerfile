FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/collector ./cmd/collector

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/collector /collector
USER nonroot:nonroot
EXPOSE 9100
ENTRYPOINT ["/collector"]
