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
	Height       = "Height"
	Tx           = "Tx"
	Elapsed      = "Elapsed"
	CommitRound  = "CommitRound"
	Round        = "Round"
	Evm          = "Evm"

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

var elapsedInfo IElapsedTimeInfos = &EmptyTimeInfo{}

func GetElapsedInfo() IElapsedTimeInfos {
	return elapsedInfo
}

var tracerObj *Tracer

func GlobalPin(format string, args ...interface{}){
	if tracerObj != nil {
		tracerObj.Pin(format, args)
	}
}

func NewTracer() *Tracer {
	tracerObj = &Tracer{
		startTime: time.Now().UnixNano(),
	}
	return tracerObj
}


type Tracer struct {
	startTime int64
	lastPin  string
	lastPinStartTime int64
	pins  []string
	intervals []int64
}


func (t *Tracer) Pin(format string, args ...interface{}) {
	t.pinByFormat(fmt.Sprintf(format, args...))
}

func (t *Tracer) pinByFormat(tag string) {
	if len(tag) == 0 {
		//panic("invalid tag")
		return
	}

	if len(t.pins) > 100 {
		// 100 pins limitation
		return
	}

	now := time.Now().UnixNano()

	if len(t.lastPin) > 0 {
		t.pins = append(t.pins, t.lastPin)
		t.intervals = append(t.intervals, (now-t.lastPinStartTime)/1e6)
	}
	t.lastPinStartTime = now
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
		info += fmt.Sprintf(", %s<%dms>", t.pins[i], t.intervals[i])
	}
	return info
}


func (t *Tracer) Reset() {
	t.startTime = time.Now().UnixNano()
	t.lastPin = ""
	t.lastPinStartTime = 0
	t.pins = nil
	t.intervals = nil
}


type EmptyTimeInfo struct {
}

func (e *EmptyTimeInfo) AddInfo(key string, info string) {
}

func (e *EmptyTimeInfo) Dump(logger log.Logger) {
}
