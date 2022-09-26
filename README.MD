# Chart-version-monitor

Chart version monitor is a relatively simple tool to monitor chart versions. It allows you to configure a bunch of
repositories and charts to monitor for changes and calls a Slack webhook to notify you when they change.

## Development
You can easily simulate a chart repository by running [http-server](https://www.npmjs.com/package/http-server) from the
project directory and renaming `example.config.json` to `config.json`.
