# Stockyard Barrage

**Load tester — define traffic scenarios, flood any URL, real-time dashboard**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted developer tools.

## Quick Start

```bash
docker run -p 9060:9060 -v barrage_data:/data ghcr.io/stockyard-dev/stockyard-barrage
```

Or with docker-compose:

```bash
docker-compose up -d
```

Open `http://localhost:9060` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9060` | HTTP port |
| `DATA_DIR` | `./data` | SQLite database directory |
| `BARRAGE_LICENSE_KEY` | *(empty)* | Pro license key |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 3 scenarios, 5 runs | Unlimited scenarios and runs |
| Price | Free | $4.99/mo |

Get a Pro license at [stockyard.dev/tools/](https://stockyard.dev/tools/).

## Category

Developer Tools

## License

Apache 2.0
