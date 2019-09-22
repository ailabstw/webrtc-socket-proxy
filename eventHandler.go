package main

import (
	"fmt"

	centrifuge "github.com/centrifugal/centrifuge-go"
	"github.com/rs/zerolog/log"
)

type subEventHandler struct {
	answerChan chan []byte
}

func (h *subEventHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	log.Print(fmt.Sprintf("New publication received from channel %s: %s", sub.Channel(), string(e.Data)))
	h.answerChan <- e.Data
}

func (h *subEventHandler) OnJoin(sub *centrifuge.Subscription, e centrifuge.JoinEvent) {
	log.Print(fmt.Sprintf("User %s (client ID %s) joined channel %s", e.User, e.Client, sub.Channel()))
}

func (h *subEventHandler) OnLeave(sub *centrifuge.Subscription, e centrifuge.LeaveEvent) {
	log.Print(fmt.Sprintf("User %s (client ID %s) left channel %s", e.User, e.Client, sub.Channel()))
}
