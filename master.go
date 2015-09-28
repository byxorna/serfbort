package main

import (
	"log"

	"github.com/hashicorp/serf/serf"
)

type Master struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func (m *Master) Run() {
	// tick every once in a while and broadcast a user message to every agent
	/*
		ticker := time.NewTicker(time.Second * 15)
		go func() {
			for t := range ticker.C {
				log.Println("Tick at", t)
				//name string, payload []byte, coalesce bool) error
				err := m.serf.UserEvent("test-event", []byte{}, false)
				if err != nil {
					log.Printf("unable to write event %s", err)
				}
			}
		}()
	*/

	for {
		select {
		case evt := <-m.eventCh:
			log.Printf("[MASTER] got event: %s", evt)
			switch evt.EventType() {
			case serf.EventMemberLeave, serf.EventMemberJoin:
				members := m.serf.Members()
				log.Printf("there are now %d members", len(members))
			}
		}
	}
}
