package trace

import (
	"fmt"
	"time"
	"github.com/tendermint/tendermint/libs/log"
)


const (
	GasUsed      = "GasUsed"
	Produce      = "Produce"
	RunTx        = "RunTx"
	Height       = "Block"
	Tx           = "Tx"
	Elapsed      = "Elapsed"
)

type IElapsedTimeInfos interface {
	AddInfo(key string, info string)
	Dump(logger log.Logger)
}

func SetInfoObject(e IElapsedTimeInfos)  {
	if e != nil {
		elapsedInfo = e
	}
}

var elapsedInfo IElapsedTimeInfos

func GetElapsedInfo() IElapsedTimeInfos {
	return elapsedInfo
}


var SleepDuration time.Duration = 0

func NewTracer() *Tracer {
	t := &Tracer{
		startTime: time.Now().UnixNano(),
	}
	return t
}


type Tracer struct {
	startTime int64
	lastPin  string
	lastTime int64
	pins  []string
	times []int64
}


func (t *Tracer) Pin(format string, args ...interface{}) {
	t.pinByFormat(fmt.Sprintf(format, args...))
}

func (t *Tracer) pinByFormat(tag string) {
	if len(tag) == 0 {
		//panic("invalid tag")
		return
	}

	now := time.Now().UnixNano()

	if len(t.lastPin) > 0 {
		t.pins = append(t.pins, t.lastPin)
		t.times = append(t.times, (now-t.lastTime)/1e6)
	}
	t.lastTime = now
	t.lastPin = tag
}


func (t *Tracer) Format() string {
	if len(t.pins) == 0 {
		return ""
	}

	t.Pin("_")

	now := time.Now().UnixNano()
	info := fmt.Sprintf("%s<%dms>",
		Elapsed,
		(now-t.startTime)/1e6,
	)
	for i := range t.pins {
		info += fmt.Sprintf(", %s<%dms>", t.pins[i], t.times[i])
	}
	return info
}


func (t *Tracer) Reset() {
	t.startTime = time.Now().UnixNano()
	t.lastPin = ""
	t.lastTime = 0
	t.pins = nil
	t.times = nil
}


