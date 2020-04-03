package log

import (
	"fmt"
	"strings"
)

type level byte

const (
	levelDebug level = 1 << iota
	levelInfo
	levelError
)

type filter struct {
	next             Logger
	allowed          level            // XOR'd levels for default case
	initiallyAllowed level            // XOR'd levels for initial case
	allowedKV        *allowedKeyvalMap // When key-value match, use this level
}

type keyval struct {
	key   interface{}
	value interface{}
}

// NewFilter wraps next and implements filtering. See the commentary on the
// Option functions for a detailed description of how to configure levels. If
// no options are provided, all leveled log events created with Debug, Info or
// Error helper methods are squelched.
func NewFilter(next Logger, options ...Option) Logger {

	allowedKV := &allowedKeyvalMap{data: make(map[keyval]level)}
	kv := keyval{"module", ""}
	allowedKV.data[kv] = levelError

	loggerMap := getLoggers()
	loggerMap.allowedKV = allowedKV


	l := &filter{
		next:      next,
		allowedKV: loggerMap.allowedKV,
	}

	for _, option := range options {
		option(l)
	}

	l.initiallyAllowed = l.allowed
	return l
}

func (l *filter) Info(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelInfo != 0
	if !levelAllowed {
		return
	}
	l.next.Info(msg, keyvals...)
}

func (l *filter) Debug(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelDebug != 0
	if !levelAllowed {
		return
	}
	l.next.Debug(msg, keyvals...)
}

func (l *filter) Error(msg string, keyvals ...interface{}) {
	levelAllowed := l.allowed&levelError != 0
	if !levelAllowed {
		return
	}
	l.next.Error(msg, keyvals...)
}

// With implements Logger by constructing a new filter with a keyvals appended
// to the logger.
//
// If custom level was set for a keyval pair using one of the
// Allow*With methods, it is used as the logger's level.
//
// Examples:
//     logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
//		 logger.With("module", "crypto").Info("Hello") # produces "I... Hello module=crypto"
//
//     logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"), log.AllowNoneWith("user", "Sam"))
//		 logger.With("module", "crypto", "user", "Sam").Info("Hello") # returns nil
//
//     logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"), log.AllowNoneWith("user", "Sam"))
//		 logger.With("user", "Sam").With("module", "crypto").Info("Hello") # produces "I... Hello module=crypto user=Sam"
func (l *filter) With(keyvals ...interface{}) Logger {
	keyInallowedKeyvalMap := false
	var keyvalsStr string
	for _, kv := range keyvals {
		s, ok := kv.(string)
		if !ok {
			return &filter{
				next:             l.next.With(keyvals...),
				allowed:          l.allowed, // simply continue with the current level
				allowedKV:        l.allowedKV,
				initiallyAllowed: l.initiallyAllowed,
			}
		}
		keyvalsStr += s
		keyvalsStr += keyvalsSplit
	}
	keyvalsStr = strings.Trim(keyvalsStr, keyvalsSplit)
	loggers := getLoggers()
	log := loggers.get(keyvalsStr)
	if log != nil {
		return log
	}

	for i := len(keyvals) - 2; i >= 0; i -= 2 {
		traverseFunc := func(kv keyval, allowed level) (bool, *filter) {
			if keyvals[i] == kv.key {
				keyInallowedKeyvalMap = true
				// Example:
				//		logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
				//		logger.With("module", "crypto")
				if keyvals[i+1] == kv.value {
					f := &filter{
						next:             l.next.With(keyvals...),
						allowed:          allowed, // set the desired level
						allowedKV:        l.allowedKV,
						initiallyAllowed: l.initiallyAllowed,
					}
					return true, f
				}
			}
			return false, nil
		}

		re, f := l.allowedKV.traverse(traverseFunc)
		if re {
			loggers.set(keyvalsStr, f)
			return f
		}
	}

	// Example:
	//		logger = log.NewFilter(logger, log.AllowError(), log.AllowInfoWith("module", "crypto"))
	//		logger.With("module", "main")
	if keyInallowedKeyvalMap {
		f := &filter{
			next:             l.next.With(keyvals...),
			allowed:          l.initiallyAllowed, // return back to initially allowed
			allowedKV:        l.allowedKV,
			initiallyAllowed: l.initiallyAllowed,
		}
		loggers.set(keyvalsStr, f)
		return f
	}

	f := &filter{
		next:             l.next.With(keyvals...),
		allowed:          l.allowed, // simply continue with the current level
		allowedKV:        l.allowedKV,
		initiallyAllowed: l.initiallyAllowed,
	}
	return f
}

//--------------------------------------------------------------------------------

// Option sets a parameter for the filter.
type Option func(*filter)

// AllowLevel returns an option for the given level or error if no option exist
// for such level.
func AllowLevel(lvl string) (Option, error) {
	switch lvl {
	case "debug":
		return AllowDebug(), nil
	case "info":
		return AllowInfo(), nil
	case "error":
		return AllowError(), nil
	case "none":
		return AllowNone(), nil
	default:
		return nil, fmt.Errorf("Expected either \"info\", \"debug\", \"error\" or \"none\" level, given %s", lvl)
	}
}

// AllowAll is an alias for AllowDebug.
func AllowAll() Option {
	return AllowDebug()
}

// AllowDebug allows error, info and debug level log events to pass.
func AllowDebug() Option {
	return allowed(levelError | levelInfo | levelDebug)
}

// AllowInfo allows error and info level log events to pass.
func AllowInfo() Option {
	return allowed(levelError | levelInfo)
}

// AllowError allows only error level log events to pass.
func AllowError() Option {
	return allowed(levelError)
}

// AllowNone allows no leveled log events to pass.
func AllowNone() Option {
	return allowed(0)
}

func allowed(allowed level) Option {
	return func(l *filter) { l.allowed = allowed }
}

// AllowDebugWith allows error, info and debug level log events to pass for a specific key value pair.
func AllowDebugWith(key interface{}, value interface{}) Option {
	return func(l *filter) {
		l.allowedKV.set(key, value, levelError|levelInfo|levelDebug)
	}
}

// AllowInfoWith allows error and info level log events to pass for a specific key value pair.
func AllowInfoWith(key interface{}, value interface{}) Option {
	return func(l *filter) {
		l.allowedKV.set(key, value, levelError|levelInfo)
	}
}

// AllowErrorWith allows only error level log events to pass for a specific key value pair.
func AllowErrorWith(key interface{}, value interface{}) Option {
	return func(l *filter) {
		l.allowedKV.set(key, value, levelError)
	}
}

// AllowNoneWith allows no leveled log events to pass for a specific key value pair.
func AllowNoneWith(key interface{}, value interface{}) Option {
	return func(l *filter) {
		l.allowedKV.set(key, value, 0)
	}
}
