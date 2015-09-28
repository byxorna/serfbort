package main

import (
	"log"

	"github.com/hashicorp/serf/serf"
)

type MasterEventHandler struct{}

func (m MasterEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		log.Printf("[EVENT] query: %s", e)
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[EVENT] user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	default:
		log.Printf("[EVENT] %s", e)
	}
}

type AgentEventHandler struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func (a AgentEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		//TODO
		log.Printf("[QUERY] implement me")
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[EVENT] user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	}
}

func (a *AgentEventHandler) EventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	for {
		select {
		case e := <-a.eventCh:
			log.Printf("[INFO] agent: Received event: %s", e.String())
			a.HandleEvent(e)

		case <-serfShutdownCh:
			log.Printf("[WARN] agent: Serf shutdown detected, quitting")
			a.serf.Shutdown()
			return

		}
	}
}
