package exec

func validateExecOptions(execOptions ExecOptions) ValidationError {
	switch execOptions.Type() {
	case ExecTypeOs:
		return nil
	default:
		return newValidationErrorUnknownExecType(execOptions.Type().String())
	}
}
