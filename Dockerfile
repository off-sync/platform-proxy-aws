# build stage
FROM golang:alpine AS build-env

WORKDIR /go/src/github.com/off-sync/platform-proxy-aws

COPY . ./

RUN go build -o /platform-proxy-aws .

# deploy stage
FROM alpine

RUN apk --no-cache add ca-certificates

COPY --from=build-env /platform-proxy-aws .

ENTRYPOINT [ "/platform-proxy-aws", "run" ]
