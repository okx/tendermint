package log

var start bool

func SetStart(st bool) {
	start = st
}

func Display() bool {
	return start
}

func SetfirstAndSecond(first int64, second int64, three uint64) {
	first1 = first
	second1 = second
	three1 = three
}

func GetfistAndSecond() (int64, int64, uint64) {
	return first1, second1, three1
}

var (
	first1  int64
	second1 int64
	three1  uint64
)
