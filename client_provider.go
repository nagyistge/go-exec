package exec

func newClientProvider(execOptions ExecOptions) (ClientProvider, error) {
	if err := validateExecOptions(execOptions); err != nil {
		return nil, err
	}
	switch execOptions.Type() {
	case ExecTypeOs:
		return newOsClientProvider(execOptions.(*OsExecOptions)), nil
	default:
		return nil, UnknownExecType(execOptions.Type())
	}
}
