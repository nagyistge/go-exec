package exec

import "sync/atomic"

const (
	volatileBoolTrue = iota
	volatileBoolFalse
)

// TODO(pedge): is this even needed? need to understand go memory model better
type volatileBool struct {
	value_ int32
}

func newVolatileBool(initialBool bool) *volatileBool {
	return &volatileBool{boolToVolatileBoolValue(initialBool)}
}

func (this *volatileBool) value() bool {
	return volatileBoolValueToBool(atomic.LoadInt32(&this.value_))
}

// return old value == new value
func (this *volatileBool) compareAndSwap(oldBool bool, newBool bool) bool {
	return atomic.CompareAndSwapInt32(
		&this.value_,
		boolToVolatileBoolValue(oldBool),
		boolToVolatileBoolValue(newBool),
	)
}

func boolToVolatileBoolValue(b bool) int32 {
	if b {
		return volatileBoolTrue
	}
	return volatileBoolFalse
}

func volatileBoolValueToBool(volatileBoolValue int32) bool {
	switch int(volatileBoolValue) {
	case volatileBoolTrue:
		return true
	case volatileBoolFalse:
		return false
	default:
		panic("exec: unknown volatileBoolValue")
	}
}
