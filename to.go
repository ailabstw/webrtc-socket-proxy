package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"

	centrifuge "github.com/centrifugal/centrifuge-go"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pion/webrtc/v2"
)

// To listen to tcp socket and proxy it to webrtc connection
type To struct {
	Listen           string
	SignalServerAddr string
	ID               string

	DataChannel *webrtc.DataChannel

	tcpConn *net.Conn
}

// NewTo returns a new "To" proxy
func NewTo(ID string, secret string, signalServerAddr string, listen string) *To {
	log.Print("new to")
	to := To{ID: ID, SignalServerAddr: signalServerAddr, Listen: listen}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": ID,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))

	signalConfig := centrifuge.DefaultConfig()
	signalClient := centrifuge.New(to.SignalServerAddr, signalConfig)
	signalClient.SetToken(tokenString)
	signalClient.Connect()

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s", connectionState.String())
	})

	webRTCConnReady := make(chan struct{})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s %d", d.Label(), d.ID())
		to.DataChannel = d

		// Register channel opening handling
		d.OnOpen(func() {
			log.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds", d.Label(), d.ID())
			webRTCConnReady <- struct{}{}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from DataChannel '%s': %#v", d.Label(), string(msg.Data))

			(*to.tcpConn).Write(msg.Data)
		})
	})
	offerChan := make(chan []byte)
	sub, err := signalClient.NewSubscription(fmt.Sprintf("offer-%s", ID))
	if err != nil {
		panic(err)
	}
	subHandler := &subEventHandler{offerChan}
	sub.OnPublish(subHandler)
	sub.OnJoin(subHandler)
	sub.OnLeave(subHandler)
	err = sub.Subscribe()
	if err != nil {
		panic(err)
	}

	history, err := sub.History()
	if err != nil {
		panic(err)
	}
	if len(history) > 0 {
		log.Printf("%s", history[0].Data)
		offerData := fmt.Sprintf("%s", history[0].Data)
		go func() { offerChan <- []byte(offerData[:]) }()
	}

	log.Print("waiting offer")
	offerData := <-offerChan
	log.Print("got offer")
	var offer webrtc.SessionDescription
	err = json.Unmarshal(offerData, &offer)
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}
	// signal answer
	answerData, err := json.Marshal(answer)
	if err != nil {
		panic(err)
	}
	err = signalClient.Publish(fmt.Sprintf("answer-%s", ID), answerData)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	return &to
}

// ListenAndServe listens on the TCP network address laddr and then handle packets
// on incoming connections.
func (s *To) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Listen)
	if err != nil {
		return err
	}
	return s.serve(listener)
}

func (s *To) serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *To) handleConn(conn net.Conn) {
	if s.tcpConn != nil {
		log.Error().Msg("Only one connection is supported")
		conn.Close()
		return
	}
	s.tcpConn = &conn
	// write to dst what it reads from src
	var pipe = func(src net.Conn) {
		defer func() {
			conn.Close()
		}()

		for {
			data := make([]byte, 1024)
			n, err := src.Read(data)
			if err != nil {
				panic(err)
			}
			log.Printf("proxying %#v", string(data[:n]))
			err = s.DataChannel.SendText(string(data[:n]))
			if err != nil {
				panic(err)
			}
		}
	}

	go pipe(conn)
}
