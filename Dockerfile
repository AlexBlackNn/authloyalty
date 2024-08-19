FROM golang:latest AS build
WORKDIR /app
COPY . /app/
RUN GOOS=linux go build -ldflags '-extldflags "-static"' -o main ./cmd/sso/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main /app/config/demo.yaml /app/
COPY --from=build  /app/boot.yaml /app/
COPY --from=build  /app/commands/proto/sso/gen /app/commands/proto/sso/gen
ENV CONFIG_PATH="/app/demo.yaml"
ENTRYPOINT ["/app/main"]


