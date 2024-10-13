FROM golang:1.23-bookworm AS go
WORKDIR /src
ADD go.mod go.sum /src/
RUN go mod download
ADD . /src
RUN go build -o /src/jimmy github.com/silas/jimmy

FROM gcr.io/distroless/base-debian12
COPY --from=go /src/jimmy /
ENTRYPOINT ["/jimmy"]
