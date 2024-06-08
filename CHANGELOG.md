# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-06-08
### Added
- add function name field option when call `WithField` in log package
- add [httpprofiling](profiling/httpprofiling) package
- add `Balancer` and  `BalancerConfig` options to event `Kafka` sender
- fix `util.Lookup` may panic on nil pointer field
- update router context, which cause unable to get http params and writer from context
- add `Increment` on cache
- monitor http counter register twice instead of grpc counter
- add `SkipRegisterReflectionServer` on grpc library
- unable to register multiple servers using grpc library, due to conflict in register reflection server
- add response total counter by path to http monitoring
- implement `http.Handler` on MyRouter
- add otel `trace_id` fields to `log.WithContext`
- remove duplicate monitor init
- add [strutil](util/strutil) package
- add `monitor.Init`, to create unique metrics per app
- add monitor metrics for grpc
- add monitor metrics for event consumer
- add `HttpRouter` method to router as replacement of `WrapperHandler`
- add `Handler` method to router to register `http.Handler`
- migrate router trace integration from opentracing to opentelemetry
- fix monitor http histogram value doesn't used the right seconds value
- removed router `WrapperHandler`, please use `router.HttpRouter` instead
- fix event consumer not using specific topic and or group config for specific topic worker pool
- fix taskworker may not work properly when the running task is panic
- adding option to override log level of kafka event library for sender and listener
- upgrade kafka-go library version to 0.4.26 because the previous version causing unexpected EOF [regression](https://github.com/segmentio/kafka-go/issues/814)
- adding stop context to event consumer to gracefully stop until context timeout
- set kafka sender and listener logger to use golib log
- set event consumer error log level to warn when error is context canceled on stop
