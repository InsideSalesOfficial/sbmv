# sbmv

Replay all dead-letter messages for a ServiceBus subscription.


## Installation

### Source
    go get github.com/InsideSalesOfficial/sbmv

## Configuration

The `SERVICEBUS_CONNECTION_STRING` environment variable must be set.


## Usage

Supply source and destination URL endpoints.

    SERVICEBUS_CONNECTION_STRING=someconnectionstring sbmv -topic some-topic -subscription some-subscription

## License

The MIT License (MIT)

Copyright (c) 2019 Brian Cozzens

See [LICENSE.md](LICENSE.md)
