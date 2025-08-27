# build
FROM golang:1.23 AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/anima .

# runtime
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/anima /app/anima
COPY docs /app/docs
ENV PORT=8081
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/app/anima"]
