FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod ./
COPY main.go ./
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/dkk .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/dkk /app/dkk

ENTRYPOINT ["/app/dkk"]
CMD ["-h"]
