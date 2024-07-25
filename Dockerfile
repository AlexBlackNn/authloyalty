FROM golang:latest AS build
WORKDIR /app
COPY . /app/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o . cmd/sso/main.go

FROM scratch
COPY --from=build /app/main /app/config/demo.yaml /app/
COPY --from=build  /app/boot.yaml /
COPY --from=build  /app/protos/proto/sso/gen /protos/proto/sso/gen
ENV CONFIG_PATH="/app/demo.yaml"
ENTRYPOINT ["/app/main"]


