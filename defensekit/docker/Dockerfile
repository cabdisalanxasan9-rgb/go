FROM golang:1.22 AS build
WORKDIR /app
COPY . .
RUN go build -o defensekit ./cmd

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /app/defensekit /app/defensekit
ENV DK_SERVE=true
ENV DK_ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 CMD ["/app/defensekit", "-healthcheck", "-health-url", "http://127.0.0.1:8080/api/health", "-silent"]
ENTRYPOINT ["/app/defensekit"]
CMD ["-serve", "-addr", ":8080"]
