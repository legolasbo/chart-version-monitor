# Chart Version Monitor

Chart Version Monitor is a relatively simple tool to monitor chart versions. It allows you to configure a bunch of
repositories and charts to monitor for changes and calls a Slack webhook to notify you when they change.

## CONFIGURATION
The container can be configured by putting a `config.json` in it's working directory or by supplying environment variables.
An example config file can be found in the repository. The supported environment variables are listed below:

* `CVM_REPOSITORIES` (*REQUIRED*) json array of repositories to monitor. See `example.config.json` to see what it should contain.
* `CVM_WEBHOOK_URL` (*REQUIRED*) the Slack webhook to call
* `CVM_REPORT_START` boolean indicating if the application should call the webhook when it starts. Defaults to true.
* `CVM_CHECK_INTERVAL` string indicating the time between checks. Must be a valid Golang duration string such as 10s, 1m10s or 1h20m30s. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w", "y". Defaults to "1h"

## Development
You can easily simulate a chart repository by running [http-server](https://www.npmjs.com/package/http-server) from the
project directory and renaming `example.config.json` to `config.json`.
