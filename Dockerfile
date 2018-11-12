FROM golang:1.11

ENV PROMETHEUS_HOST "http://prometheus.istio-system.svc.cluster.local:9090"
ENV WEB_DIST_VERSION "v1.0.0-alpha.9"
ENV WEB_DIST_PATH "/go/src/app/web-dist"
ENV GIN_MODE "release"

# Force the go compiler to use modules
ENV GO111MODULE=on

WORKDIR /go/src/app

# Download frontend
RUN curl https://s3-us-west-1.amazonaws.com/outlier-cloud-us-west-1/web-dist/${WEB_DIST_VERSION}.tar.gz --output web-dist.tar.gz \
        && tar -xzf web-dist.tar.gz \
        && mv ./build ./web-dist \
        && rm web-dist.tar.gz

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build

CMD ["./outlier-istio"]
