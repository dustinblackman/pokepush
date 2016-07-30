# Pokepush

[![Build Status](https://travis-ci.org/dustinblackman/pokepush.svg?branch=master)](https://travis-ci.org/dustinblackman/pokepush)
[![Go Report Card](https://goreportcard.com/badge/github.com/dustinblackman/pokepush)](https://goreportcard.com/report/github.com/dustinblackman/pokepush)

Dead simple Pokemon Go push notifications through Pushbullet when a new Pokemon arrives in a specified locations. Supports multiple users subscribed to desired locations.

## Setup

- Download the binaries for your system from [releases](https://github.com/dustinblackman/pokepush/releases).
- Acquire a pushbullet token from your accounts [settings page](https://www.pushbullet.com/#settings).
- Create config.json next to the binary.
- Start.

**Example Config:**
```json
{
  "users": [
    {
      "name": "Jim",
      "pushbullet_key": "PUSHBULLET_KEY",
      "places": ["Home", "Work"]
    }
  ],
  "places": [
    {"name": "Home", "lat": 51.501476, "lon": -0.140634},
    {"name": "Work", "lat": 51.5159394, "lon": -0.1352864}
  ],
  "logs": "pretty"
}
```

- `users` - An array of users containing their `name`, `pushbullet_key`, an array of `places`. Names of places must match the ones in the `places` array.
- `places` - An array of locations with their `name`, `lat`, and `lon`. You can covert an address to latitude/longitude [here](http://www.latlong.net/convert-address-to-lat-long.html)
- `logs` - Sets log type, supports `pretty` and `json`. *Default is `pretty`.*

## Docker

A docker image is also available over at [Docker Hub](https://hub.docker.com/r/dustinblackman/pokepush/).

## Development

After git pulling the repo in to your `$GOPATH`, executing the following to get a development setup running.

```bash
make deps
make dev
```

## TODO
- Handle errors
- Write tests
- Handle scheduled places to only push notifications during specified times of day

## Credit
- [PokeVision](http://pokevision)

## [License](LICENSE)
