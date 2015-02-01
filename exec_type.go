package exec

import (
	"errors"
	"fmt"
)

var (
	ExecTypeOs ExecType = 0

	execTypeToString = map[ExecType]string{
		ExecTypeOs: "os",
	}
	lenExecTypeToString = len(execTypeToString)
	stringToExecType    = map[string]ExecType{
		"os": ExecTypeOs,
	}
)

type ExecType uint

func validExecType(s string) bool {
	_, ok := stringToExecType[s]
	return ok
}

func execTypeOf(s string) (ExecType, error) {
	execType, ok := stringToExecType[s]
	if !ok {
		return 0, errors.New(unknownExecType(s))
	}
	return execType, nil
}

func (this ExecType) string() string {
	if int(this) < lenExecTypeToString {
		return execTypeToString[this]
	}
	panic(unknownExecType(this))
}

func unknownExecType(unknownExecType interface{}) string {
	return fmt.Sprintf("exec: unknown ExecType: %v", unknownExecType)
}
