FROM golang:1.26.4 AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /validator ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /validator /validator
USER nonroot:nonroot
ENTRYPOINT ["/validator"]
