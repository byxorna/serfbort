package main

import (
	"log"

	"github.com/hashicorp/serf/serf"
)

type MasterEventHandler struct{}

func (m MasterEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("Got user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	}
}
