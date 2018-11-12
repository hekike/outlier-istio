# outlier-istio

[![CircleCI](https://circleci.com/gh/hekike/outlier-istio.svg?style=svg&circle-token=5504d6f60f8b1217aadf87342112f6d29ccaac2a)](https://circleci.com/gh/hekike/outlier-istio)

Root cause analysis tool for Kubernetes Istio.

Outlier detection with approximate median latency for Istio workloads.

![outlier-istio workloads](https://user-images.githubusercontent.com/1764512/48328025-e1559d00-e5f6-11e8-8129-15ff9c003554.png)

![outlier-istio workload](https://user-images.githubusercontent.com/1764512/48328024-e1559d00-e5f6-11e8-8a31-237ad0afad9a.png)

## Components

- Frontend: https://github.com/hekike/outlier-web

## Install

**Requirements**

- Istio (with Prometheus)

```sh
kubectl create -f ./install/deployment.yaml
```

**Test**

```sh
kubectl -n istio-system port-forward deployment/outlier-istio 8080
open http://localhost:8080
```

## API

Inlined OpenAPI (Swagger).

### Requirements

https://goswagger.io

```sh
brew tap go-swagger/go-swagger
brew install go-swagger
```

### Generate

```sh
swagger generate spec -o swagger.json
swagger serve swagger.json
```
