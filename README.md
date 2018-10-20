# outlier-istio

[![CircleCI](https://circleci.com/gh/hekike/outlier-istio.svg?style=svg&circle-token=5504d6f60f8b1217aadf87342112f6d29ccaac2a)](https://circleci.com/gh/hekike/outlier-istio)

TODO

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
