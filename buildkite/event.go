package buildkite

import (
	"encoding/json"
	"fmt"

	bk "github.com/buildkite/go-buildkite/buildkite"
)

type EventType string

const (
	EventTypePing              EventType = "ping"
	EventTypeBuildScheduled              = "build.scheduled"
	EventTypeBuildRunning                = "build.running"
	EventTypeBuildFinished               = "build.finished"
	EventTypeJobStarted                  = "job.started"
	EventTypeJobFinished                 = "job.finished"
	EventTypeJobActivated                = "job.activated"
	EventTypeAgentConnected              = "agent.connected"
	EventTypeAgentLost                   = "agent.lost"
	EventTypeAgentDisconnected           = "agent.disconnected"
	EventTypeAgentStopping               = "agent.stopping"
	EventTypeAgentStopped                = "agent.stopped"
)

func Unmarshal(b []byte) (Event, error) {
	var intermediate struct {
		Event EventType
	}
	err := json.Unmarshal(b, &intermediate)
	if err != nil {
		return nil, err
	}

	switch intermediate.Event {
	case EventTypePing:
		pe := PingEvent{}
		err := json.Unmarshal(b, &pe)
		return pe, err
	case EventTypeBuildScheduled, EventTypeBuildRunning, EventTypeBuildFinished:
		be := BuildEvent{}
		err := json.Unmarshal(b, &be)
		return be, err
	case EventTypeJobActivated, EventTypeJobStarted, EventTypeJobFinished:
		je := JobEvent{}
		err := json.Unmarshal(b, &je)
		return je, err
	case EventTypeAgentConnected, EventTypeAgentDisconnected, EventTypeAgentStopped, EventTypeAgentStopping, EventTypeAgentLost:
		ae := AgentEvent{}
		err := json.Unmarshal(b, &ae)
		return ae, err
	default:
		return nil, fmt.Errorf("unrecognized event: %v", intermediate.Event)
	}
}

type PingEvent struct {
	Event   EventType
	Service struct {
		ID       string `json:"id"`
		Provider string
		Settings map[string]string
	}
	Organization bk.Organization
	Sender       EventSender
}

type BuildEvent struct {
	Event    EventType
	Build    bk.Build
	Pipeline bk.Pipeline
	Sender   EventSender
}

type JobEvent struct {
	Event    EventType
	Job      bk.Job
	Build    bk.Build
	Pipeline bk.Pipeline
	Sender   EventSender
}

type AgentEvent struct {
	Event  EventType
	Agent  bk.Agent
	Sender EventSender
}

type EventSender struct {
	ID   string `json:"id"`
	Name string
}

type Event interface {
	Type() EventType

	// cannot be implemented outside of this package
	closed()
}

func (pe PingEvent) Type() EventType {
	return EventTypePing
}
func (pe PingEvent) closed() {}

func (be BuildEvent) Type() EventType { return be.Event }
func (be BuildEvent) closed()         {}

func (je JobEvent) Type() EventType { return je.Event }
func (je JobEvent) closed()         {}

func (ae AgentEvent) Type() EventType { return ae.Event }
func (ae AgentEvent) closed()         {}
