## 2.0.3 16 Apr 2025
- Fix standard_logger.go file and line number reference on all log messages

## 2.0.2 30 Jun 2023

- Add Fatal log level
- Update the Sentry logger to only create events of level Warning and above, and add messages below Warning as breadcrumbs

## 2.0.1 31 May 2023

- No code changes; version bump to fix moved v2.0.0 commit

## 2.0.0 31 May 2023

- Refactored support for multiple log sinks, and added support for Sentry logging
- All loggers accept a list of user properties to log (e.g., email, user ID)
- Breaking changes:
- Service configuration now requires a `ServiceName` property, which is logged as a Sentry event tag
- Log functions have been moved from `service` to the `log` package
- Logs of the form `service.Log.[Level](...)` must be moved to `log.Log(log.[Level], ...)`
- The Firebase client constructor signature has changed, and now explicitly takes a firebase config struct and credentials file
- Add required `serviceName` field to base configuration

## 1.0.8 14 Feb 2023

- Everything that came before
