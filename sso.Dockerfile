FROM golang:latest AS build
WORKDIR /app
COPY . /app/
# go build -ldflags "-s -w" — скомпилирует исполняемый файл меньшего размера, так как в него не будет включена таблица символов и отладочная информация.
RUN GOOS=linux go build -ldflags '-extldflags "-static" -s -w' -o main ./sso/cmd/sso/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main /app/sso/config/demo.yaml /app/
COPY --from=build  /app/commands/proto/sso/gen /app/commands/proto/sso/gen
ENV CONFIG_PATH="/app/demo.yaml"
ENTRYPOINT ["/app/main"]


