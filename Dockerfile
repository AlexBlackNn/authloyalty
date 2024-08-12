FROM golang:latest AS build
WORKDIR /app
COPY . /app/
RUN GOOS=linux go build -a -o . cmd/sso/main.go

FROM golang:latest
WORKDIR /app
COPY --from=build /app/main /app/config/demo.yaml /app/
COPY --from=build  /app/boot.yaml /app/
COPY --from=build  /app/protos/proto/sso/gen /app/protos/proto/sso/gen
ENV CONFIG_PATH="/app/demo.yaml"
RUN ls -la
ENTRYPOINT ["/app/main"]


