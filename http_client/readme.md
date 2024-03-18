# `http_client` package

# Objective
the purpose of this package is to standardized 
http request for each services

# Functionality
3 main features:

1. Basic HTTP Client request
2. HTTP Client request with timeout (featuring Netflix's Hystrix)
3. HTTP Client request with singleflight

# Technical Depth

The `http_client` package uses the [decorator pattern](https://refactoring.guru/design-patterns/decorator).

Under the hood, the `http_client` package depends on:

1. https://pkg.go.dev/net/http
2. https://pkg.go.dev/golang.org/x/sync/singleflight
3. https://github.com/afex/hystrix-go

# Examples
Find the `examples` folder for references
