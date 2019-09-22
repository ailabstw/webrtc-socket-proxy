# webrtc-socket-proxy

![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)
![travis-ci](https://travis-ci.org/poga/webrtc-socket-proxy.svg?branch=master)

Peer-to-peer TCP socket proxy with WebRTC, use [centrifugo](https://centrifugal.github.io/centrifugo/) as the signal server.

## Setup

* Install `webrtc-socket-proxy`

```
$ go get -u github.com/poga/webrtc-socket-proxy
```

* Setup [centrifugo](https://github.com/centrifugal/centrifugo/releases) with [example config](config.centrifugo.test.json)>

* Start proxies

```
$ webrtc-socket-proxy -signal=<SIGNAL_SERVER_ADDR> -secret=<SIGNAL_SERVER_SECRET> -as=<PEER_ID> -upstreamAddr=localhost:8000
$ webrtc-socket-proxy -signal=<SIGNAL_SERVER_ADDR> -secret=<SIGNAL_SERVER_SECRET> -to=<PEER_ID> -listen=:4444
```


## Usage


## Roadmap

- [ ] Multiplex Connections. Currently we only support one connnection per proxy

## License

The MIT License