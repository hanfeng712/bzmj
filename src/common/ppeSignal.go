package common

import "fmt"
import "os"
import "os/signal"

type SignalHandler func(s os.Signal, arg interface{})

type signalSet struct {
	m map[os.Signal]SignalHandler
}

func signalSetNew() *signalSet {
	ss := new(signalSet)
	ss.m = make(map[os.Signal]SignalHandler)
	return ss
}

func (set *signalSet) register(s os.Signal, handler SignalHandler) {
	if _, found := set.m[s]; !found {
		set.m[s] = handler
	}
}

func (set *signalSet) handle(sig os.Signal, arg interface{}) (err error) {
	if _, found := set.m[sig]; found {
		set.m[sig](sig, arg)
		return nil
	} else {
		return fmt.Errorf("No handler available for signal %v", sig)
	}

	panic("won't reach here")
}

func WatchSystemSignal(watchsingals *[]os.Signal, callbackHandler SignalHandler) {
	ss := signalSetNew()

	for _, wathsingnal := range *watchsingals {
		ss.register(wathsingnal, callbackHandler)
	}

	for {
		c := make(chan os.Signal)
		var sigs []os.Signal
		for sig := range ss.m {
			sigs = append(sigs, sig)
		}
		signal.Notify(c)
		sig := <-c

		err := ss.handle(sig, nil)
		if err != nil {
			fmt.Printf("unknown signal received: %v\n", sig)
		}
	}
}
