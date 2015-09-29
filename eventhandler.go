package main

import (
	"log"
	"sync"

	"github.com/hashicorp/serf/serf"
)

type MasterEventHandler struct {
	Queries []*serf.Query
	sync.Mutex
}

func (m *MasterEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		query := e.(*serf.Query)
		log.Printf("[EVENT] %s", query)
		m.Lock()
		defer m.Unlock()
		m.Queries = append(m.Queries, query)
		//TODO broadcast this query? it came in over RPC...
		log.Print("TODO broadcast this query!")
	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[EVENT] user event %s with payload %q (coalescable: %t)", ue.Name, ue.Payload, ue.Coalesce)
	default:
		log.Printf("[EVENT] %s", e)
	}
}

/****
//---------------------- RIP ME OUT
// AgentQueryHandler is an serf.EventHandler implementation
type AgentQueryHandler struct {
	Response []byte
	Queries  []*serf.Query
	sync.Mutex
}

func (h *AgentQueryHandler) HandleEvent(e serf.Event) {
	query, ok := e.(*serf.Query)
	if !ok {
		return
	}

	h.Lock()
	h.Queries = append(h.Queries, query)
	h.Unlock()

	query.Respond(h.Response)
}

//---------------------- END RIP ME OUT
***/

type AgentEventHandler struct {
	serf    *serf.Serf
	eventCh chan serf.Event
}

func (a AgentEventHandler) HandleEvent(e serf.Event) {
	switch e.EventType() {
	case serf.EventQuery:
		log.Print("[QUERY] received a query")
		query := e.(*serf.Query)
		//TODO track this query!
		log.Printf("[QUERY] received a query %v", query)
		payload, err := decodeMessagePayload(query.Payload)
		if err != nil {
			log.Printf("[ERROR] unable to decode payload: %s", err)
			return
		}
		log.Printf("[QUERY] parsed message: %s", payload)
		err = query.Respond([]byte("Hey this is a response"))
		if err != nil {
			log.Printf("[ERROR] unable to respond to query: %s", err)
			return
		}

	case serf.EventUser:
		ue := e.(serf.UserEvent)
		log.Printf("[TESTING] %v", ue)
		switch ue.Name {
		case "deploy":
			log.Printf("[DEPLOY] received payload %q (coalescable: %t)", ue.Payload, ue.Coalesce)
			messagePayload, err := decodeMessagePayload(ue.Payload)
			if err != nil {
				log.Printf("[ERROR] unable to decode payload: %s", err)
				return
			}
			log.Printf("[DEPLOY] parsed deploy message: %s", messagePayload)

			target, ok := config.Targets[messagePayload.Target]
			if !ok {
				log.Printf("[ERROR] No target configured named %q", messagePayload.Target)
				return
			}

			log.Printf("[DEPLOY] target %s with message %q target %s", messagePayload.Target, messagePayload, target)
			//TODO FIXME do something here...
		case "verify":
			//TODO implement me
		default:
			log.Printf("[WARN] unknown message received: %s with payload %q", ue.Name, ue.Payload)
		}
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
