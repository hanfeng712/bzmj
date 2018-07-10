package stats

import (
	"bytes"
	"expvar"
	"fmt"
	"sync"
	"time"
)

type States struct {
	// set at construction time
	StateCount int
	Labels     []string

	// the following variables can change, protected by mutex
	mu                    sync.Mutex
	CurrentState          int
	CurrentStateStartTime time.Time // when we switched to our state

	// historical data about the states
	Durations   []time.Duration // how much time in each state
	Transitions []int           // how many times we got into a state
}

func NewStates(name string, labels []string, startTime time.Time, initialState int) *States {
	s := &States{StateCount: len(labels), Labels: labels, CurrentState: initialState, CurrentStateStartTime: startTime, Durations: make([]time.Duration, len(labels)), Transitions: make([]int, len(labels))}
	if initialState < 0 || initialState >= s.StateCount {
		panic(fmt.Errorf("initialState out of range 0-%v: %v", s.StateCount, initialState))
	}
	if name != "" {
		expvar.Publish(name, s)
	}
	return s
}

func (s *States) SetState(state int) {
	s.SetStateAt(state, time.Now())
}

func (s *States) SetStateAt(state int, now time.Time) {
	if state < 0 || state >= s.StateCount {
		panic(fmt.Errorf("State out of range 0-%v: %v", s.StateCount, state))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// check we're going strictly forward in time
	dur := now.Sub(s.CurrentStateStartTime)
	if dur < 0 {
		panic(fmt.Errorf("Time going backwards? %v < %v", now, s.CurrentStateStartTime))
	}

	// record the previous state duration, reset our state
	s.Durations[s.CurrentState] += dur
	s.Transitions[state] += 1
	s.CurrentState = state
	s.CurrentStateStartTime = now
}

func (s *States) String() string {
	return s.StringAt(time.Now())
}

func (s *States) StringAt(now time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := bytes.NewBuffer(make([]byte, 0, 4096))
	fmt.Fprintf(b, "{")

	fmt.Fprintf(b, "\"Current\": \"%v\"", s.Labels[s.CurrentState])

	for i := 0; i < s.StateCount; i++ {

		d := s.Durations[i]
		t := s.Transitions[i]
		if i == s.CurrentState {
			dur := now.Sub(s.CurrentStateStartTime)
			if dur > 0 {
				d += dur
			}
		}

		fmt.Fprintf(b, ", ")
		fmt.Fprintf(b, "\"Duration%v\": %v, ", s.Labels[i], int64(d))
		fmt.Fprintf(b, "\"TransitionInto%v\": %v", s.Labels[i], t)
	}

	fmt.Fprintf(b, "}")
	return b.String()
}
