FROM golang:latest AS build
WORKDIR /application
COPY . /application/
RUN GOOS=linux go build -a -o . cmd/sso/main.go
RUN ls -la
ENV CONFIG_PATH="/application/config/demo.yaml"
ENTRYPOINT ["/application/main"]



