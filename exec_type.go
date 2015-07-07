package exec

import "fmt"

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

func AllExecTypes() []ExecType {
	return []ExecType{
		ExecTypeOs,
	}
}

func ExecTypeOf(s string) (ExecType, error) {
	execType, ok := stringToExecType[s]
	if !ok {
		return 0, UnknownExecType(s)
	}
	return execType, nil
}

func (e ExecType) String() string {
	if int(e) < lenExecTypeToString {
		return execTypeToString[e]
	}
	panic(UnknownExecType(e).Error())
}

func UnknownExecType(unknownExecType interface{}) error {
	return fmt.Errorf("exec: unknown ExecType: %v", unknownExecType)
}
