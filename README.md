# WUERFLER

A small app to roll dices with your party. Inspiration taken from http://rolldicewithfriends.com/

## Requirements

- go
- yarn

## Building

```
(cd frontend && yarn)
make
```

This will result in a `wuerfler` binary and a `frontend-build` directory.

## Running

`sudo ./wuerfler # default port 80 so sudo required`

There are a few environment variables to control wuerfler:

WUERFLER_PORT=80 HTTP Port
WUERFLER_SECUREPORT= HTTPS Port. Also needs WUERFLER_SECUREHOSTNAME
WUERFLER_SECUREHOSTNAME=example.com

## Local development

`WUERFLER_DEBUG=1 WUERFLER_PORT=3000 go run main.go`

In a separate window:

`cd frontend && yarn watch`
