FROM golang:1.16 AS build-env

ENV PKG_PATH /go/src/github.com/nomuyoshi/grpc-samples/todo-microservices
ENV SERVICE_PATH $PKG_PATH/front
WORKDIR $PKG_PATH
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . $PKG_PATH
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install $SERVICE_PATH

FROM alpine:3.13
ENV TZ Asia/Tokyo
COPY --from=build-env /go/bin/front /go/bin/front
COPY front/static /static
COPY front/template /template
ENTRYPOINT ["/go/bin/front"]
