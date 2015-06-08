package main

import (
	"strings"
	"time"
)

const (
	STATE_START StateEnum = iota
	STATE_NOTIFY1
	STATE_NOTIFY2
	STATE_REAPABLE
	STATE_IGNORE
	STATE_WHITELIST
)

type StateEnum int

// StateEnum.String() in stateenum_string.go

type State struct {
	State StateEnum

	// State must be maintained until this time
	Until time.Time
}

func (s *State) String() string {
	return s.State.String() + s_sep + s.Until.Format(s_tformat)
}

func ParseState(state string) (defaultState *State) {

	defaultState = &State{STATE_START, time.Time{}}

	if state == "" {
		return
	}

	s := strings.Split(state, s_sep)

	if len(s) != 2 {
		return
	}

	var stateEnum StateEnum
	switch s[0] {
	case "STATE_START":
		stateEnum = STATE_START
	case "STATE_NOTIFY1":
		stateEnum = STATE_NOTIFY1
	case "STATE_NOTIFY2":
		stateEnum = STATE_NOTIFY2
	case "STATE_IGNORE":
		stateEnum = STATE_IGNORE
	case "STATE_WHITELIST":
		stateEnum = STATE_WHITELIST
	case "STATE_REAPABLE":
		stateEnum = STATE_REAPABLE
	default:
		return
	}

	t, err := time.Parse(s_tformat, s[1])
	if err != nil {
		return
	}

	return &State{stateEnum, t}
}
