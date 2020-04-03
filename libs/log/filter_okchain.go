package log

import (
	"fmt"
	"strings"
	"sync"
)

const (
	keyvalsSplit = "&"
)

type allowedKeyvalMap struct {
	sync.RWMutex
	data map[keyval]level // When key-value match, use this level
}

func (a *allowedKeyvalMap) set(key interface{}, value interface{}, lv level) {
	a.Lock()
	defer a.Unlock()
	a.data[keyval{key, value}] = lv
}

func (a *allowedKeyvalMap) traverse(f func(keyval, level) (bool, *filter)) (bool, *filter) {
	a.RLock()
	defer a.RUnlock()
	for kv, allowed := range a.data {
		re, f := f(kv, allowed)
		if re {
			return re, f
		}
	}
	return false, nil
}

type CacheLoggers struct {
	sync.RWMutex
	allowedKV  *allowedKeyvalMap
	loggersMap map[string]Logger
}

var once sync.Once
var cacheLoggers *CacheLoggers

func (cl *CacheLoggers) update(defaultOption Option, options ...Option) {

	cl.Lock()
	defer cl.Unlock()

	for _, option := range options {
		option(&filter{allowedKV: cl.allowedKV})
	}

	for k, v := range cl.loggersMap {
		l, ok := v.(*filter)
		if !ok {
			continue
		}

		if defaultOption != nil {
			defaultOption(l)
		}

		l.initiallyAllowed = l.allowed // allowed: default * allowed
		ks := strings.Split(k, keyvalsSplit)
		l.UpdateWith(ks...)
	}
}

func UpdateFilter(defaultOption Option, options ...Option) {
	loggers := getLoggers()
	loggers.update(defaultOption, options...)
}

func (l *CacheLoggers) get(key string) Logger {
	l.RLock()
	defer l.RUnlock()
	if value, ok := l.loggersMap[key]; ok {
		return value
	}
	return nil
}

func (l *CacheLoggers) set(key string, logger Logger) {
	l.Lock()
	defer l.Unlock()
	l.loggersMap[key] = logger
}

func getLoggers() *CacheLoggers {
	once.Do(func() {
		cacheLoggers = &CacheLoggers{
			loggersMap: make(map[string]Logger),
		}
	})
	return cacheLoggers
}

func (l *filter) UpdateWith(keyvals ...string) {
	keyInallowedKeyvalMap := false

	for i := len(keyvals) - 2; i >= 0; i -= 2 {

		traverseFunc := func(kv keyval, allowed level) (bool, *filter) {
			if keyvals[i] != kv.key {
				return false, nil
			}

			keyInallowedKeyvalMap = true

			if keyvals[i+1] != kv.value {
				return false, nil
			}
			l.allowed = allowed // set the desired level
			return true, nil
		}

		re, _ := l.allowedKV.traverse(traverseFunc)
		if re {
			return
		}
	}

	if keyInallowedKeyvalMap {
		l.allowed = l.initiallyAllowed // return back to initially allowed
	}
}

func UpdateLogLevel(lvl string) error {
	if lvl == "" {
		return fmt.Errorf("Empty log level")
	}

	defaultLogLevelKey := "*"
	l := lvl

	// prefix simple one word levels (e.g. "info") with "*"
	if !strings.Contains(l, ":") {
		l = defaultLogLevelKey + ":" + l
	}

	options := make([]Option, 0)

	var defaultOption Option // for module *
	var err error

	list := strings.Split(l, ",")
	for _, item := range list {
		moduleAndLevel := strings.Split(item, ":")

		if len(moduleAndLevel) != 2 {
			return fmt.Errorf("Expected list in a form of \"module:level\" pairs, given pair %s, list %s", item, list)
		}

		module := moduleAndLevel[0]
		level := moduleAndLevel[1]

		var option Option
		if module == defaultLogLevelKey {
			defaultOption, err = AllowLevel(level)
			if err != nil {
				return err
			}
		} else {
			switch level {
			case "debug":
				option = AllowDebugWith("module", module)
			case "info":
				option = AllowInfoWith("module", module)
			case "error":
				option = AllowErrorWith("module", module)
			case "none":
				option = AllowNoneWith("module", module)
			default:
				return fmt.Errorf("Expected either \"info\", \"debug\", \"error\" or \"none\" log level, given %s (pair %s, list %s)",
					level, item, list)
			}
			options = append(options, option)
		}
	}

	UpdateFilter(defaultOption, options...)
	return nil
}
