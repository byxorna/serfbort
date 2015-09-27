package main

import (
	"log"

	"github.com/hashicorp/serf/serf"
)

type Slave struct {
	serf    *serf.Serf
	eventCh <-chan serf.Event
}

func (s *Slave) Run() {
	for {
		select {
		case evt := <-s.eventCh:
			log.Printf("[SLAVE] got event: %s", evt)
			switch evt.EventType() {
			case serf.EventUser:
				ue := evt.(serf.UserEvent)
				log.Printf("user event: %s", ue.String())
			}
		}
	}
}
