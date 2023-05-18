## 2.0.0 18 May 2023

- Refactored support for multiple log sinks, and added support for Sentry logging
- Breaking changes:
- Log functions have been moved from `service` to the `log` package
- Logs of the form `service.Log.[Level](...)` must be moved to `log.Log(log.[Level], ...)`
- The Firebase client constructor signature has changed, and now explicitly takes a firebase config struct and credentials file

## 1.0.8 14 Feb 2023

- Everything that came before
