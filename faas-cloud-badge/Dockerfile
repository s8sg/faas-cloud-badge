FROM golang:1.9.4 as build

WORKDIR /go/src/github.com/faasinating/faas-cloud-badge/faas-cloud-badge

COPY vendor         vendor
COPY server.go      .
COPY commitstatus.go .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o faasbadge .

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /home/app

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /go/src/github.com/faasinating/faas-cloud-badge/faas-cloud-badge/faasbadge    .
COPY assets  assets

RUN chown -R app:app ./

USER app

CMD ["./faasbadge"]
