package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/pion/webrtc/v2"
	"github.com/rs/zerolog/log"

	centrifuge "github.com/centrifugal/centrifuge-go"
	jwt "github.com/dgrijalva/jwt-go"
)

// As listen to new webRTC connection and proxy it to upstream tcp socket
type As struct {
	UpstreamAddr     string
	SignalServerAddr string
	ID               string

	peerConn *webrtc.PeerConnection

	UpstreamChan chan []byte
}

// NewAs returns a new "As" proxy
func NewAs(ID string, secret string, upstreamAddr string, signalServerAddr string) *As {
	log.Print("hello")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": ID,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))

	as := As{ID: ID, UpstreamAddr: upstreamAddr, SignalServerAddr: signalServerAddr}
	upChan := make(chan []byte)
	as.UpstreamChan = upChan

	signalConfig := centrifuge.DefaultConfig()
	log.Print(as.SignalServerAddr)
	signalClient := centrifuge.New(as.SignalServerAddr, signalConfig)
	signalClient.SetToken(tokenString)
	signalClient.Connect()

	webRTCconfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(webRTCconfig)
	if err != nil {
		panic(err)
	}

	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		panic(err)
	}

	dataChannel.OnOpen(func() {
		log.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds", dataChannel.Label(), dataChannel.ID())

		if as.UpstreamAddr == "" {
			log.Error().Msgf("cannot connect to upstream \"%s\"", as.UpstreamAddr)
			return
		}
		go func(upChan chan []byte) {
			upAddr, err := net.ResolveTCPAddr("tcp", as.UpstreamAddr)
			if err != nil {
				panic(err)
			}
			upConn, err := net.DialTCP("tcp", nil, upAddr)
			if err != nil {
				panic(err)
			}
			// read from socket
			go func() {
				for {
					data := make([]byte, 1024)
					n, err := upConn.Read(data)
					if err != nil {
						panic(err)
					}
					log.Printf("sending to down stream %#v", string(data[:n]))
					dataChannel.SendText(string(data[:n]))
				}
			}()

			for {
				data := <-upChan

				if _, err := upConn.Write(data); err != nil {
					panic(err)
				}
			}
		}(upChan)
	})
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("Message from DataChannel '%s': %#v", dataChannel.Label(), string(msg.Data))

		as.UpstreamChan <- msg.Data
	})

	// Create an offer to send to the browser
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	offerData, err := json.Marshal(offer)
	if err != nil {
		panic(err)
	}
	err = signalClient.Publish(fmt.Sprintf("offer-%s", ID), offerData)
	if err != nil {
		panic(err)
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	answerChan := make(chan []byte)
	sub, err := signalClient.NewSubscription(fmt.Sprintf("answer-%s", ID))
	if err != nil {
		panic(err)
	}
	subHandler := &subEventHandler{answerChan}
	sub.OnPublish(subHandler)
	sub.OnJoin(subHandler)
	sub.OnLeave(subHandler)

	err = sub.Subscribe()
	if err != nil {
		panic(err)
	}

	log.Print("waiting answer")

	// wait answer
	answerData := <-answerChan
	log.Print("got answer")
	var answer webrtc.SessionDescription
	err = json.Unmarshal(answerData, &answer)
	if err != nil {
		panic(err)
	}

	log.Printf("%v", answer)

	// Apply the answer as the remote description
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	return &as
}
