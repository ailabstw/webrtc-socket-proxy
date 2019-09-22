package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	listenAddr       = flag.String("listen", ":4444", "proxy listen address")
	upstreamAddr     = flag.String("upstream", "", "proxy upstream address")
	signalServerAddr = flag.String("signal", "ws://localhost:8000/connection/websocket", "signaling server address")

	secret = flag.String("secret", "", "server secret")

	as = flag.String("as", "", "proxy ID")
	to = flag.String("to", "", "proxy target ID")
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	flag.Parse()

	if *as != "" {
		NewAs(*as, *secret, *upstreamAddr, *signalServerAddr)

		for {
		}
	} else if *to != "" {
		p := NewTo(*to, *secret, *signalServerAddr, *listenAddr)

		log.Print("Proxy listening" + p.Listen + " to target " + p.ID)
		p.ListenAndServe()

		return
	}
}
