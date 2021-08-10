package state

import (
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"time"
)

var IgnoreSmbCheck bool = false

type Tracer struct {
	startTime int64

	lastPin  string
	lastTime int64

	pins  []string
	times []int64
	l     log.Logger
}

func NewTracer(l log.Logger) *Tracer {
	return &Tracer{l: l.With("module", "main")}
}

func (t *Tracer) pin(tag string) {
	if len(tag) == 0 {
		//panic("invalid tag")
		return
	}

	now := time.Now().UnixNano()

	if t.startTime == 0 {
		t.startTime = now
	}

	if len(t.lastPin) > 0 {
		t.pins = append(t.pins, t.lastPin)
		t.times = append(t.times, (now-t.lastTime)/1e6)
	}
	t.lastTime = now
	t.lastPin = tag
}

func (t *Tracer) dump(caller string) {
	t.pin("_")
	now := time.Now().UnixNano()

	dumpRet := caller
	for i := range t.pins {
		dumpRet += fmt.Sprintf("%s=<%dms>, ", t.pins[i], t.times[i])
	}

	t.l.Info(dumpRet, "ElapsedTime=", (now-t.startTime)/1e6)
}
