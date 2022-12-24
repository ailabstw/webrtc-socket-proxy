<h1 align="center">
  webrtc-socket-proxy
</h1>
<h4 align="center">Seamless peer-to-peer TCP socket proxy using WebRTC, with <a href="https://centrifugal.github.io/centrifugo/">centrifugo</a> as the signal server</h4>

<p align="center">
  <img src="https://img.shields.io/badge/stability-experimental-orange.svg">
  <img src="https://travis-ci.org/ailabstw/webrtc-socket-proxy.svg?branch=master">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a><br />
  <br />
  <img src="./how.png"><br/>
</p>

## Setup

* Install `webrtc-socket-proxy` on both peers

```
$ go get -u github.com/poga/webrtc-socket-proxy
```

* On the third machine with a dedicated IP, setup [centrifugo](https://github.com/centrifugal/centrifugo/releases) with [example config](config.centrifugo.test.json).

## Usage

```
# the `As` proxy
$ webrtc-socket-proxy -signal=<SIGNAL_SERVER_ADDR> -secret=<SIGNAL_SERVER_SECRET> -as=<PEER_ID> -upstreamAddr=localhost:8000
# the `To` proxy
$ webrtc-socket-proxy -signal=<SIGNAL_SERVER_ADDR> -secret=<SIGNAL_SERVER_SECRET> -to=<PEER_ID> -listen=:4444
```

You can send data from the `As` machine to your `<upstreamAddr>` via connecting to `:4444` of the `To` machine now.

## Roadmap

- [ ] TURN server support
- [ ] Multiplex Connections. Currently we only support one connnection per proxy-pair

## License

The MIT License
