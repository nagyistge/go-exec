package exec

func newExternalClientProvider(externalExecOptions *ExternalExecOptions) (ClientProvider, error) {
	execOptions, err := ConvertExternalExecOptions(externalExecOptions)
	if err != nil {
		return nil, err
	}
	return newClientProvider(execOptions)
}

func ConvertExternalExecOptions(externalExecOptions *ExternalExecOptions) (ExecOptions, error) {
	if !validExecType(externalExecOptions.Type) {
		return nil, newValidationErrorUnknownExecType(externalExecOptions.Type)
	}
	execType, err := execTypeOf(externalExecOptions.Type)
	if err != nil {
		return nil, err
	}
	switch execType {
	case ExecTypeOs:
		return &OsExecOptions{
			TmpDir: externalExecOptions.TmpDir,
		}, nil
	default:
		return nil, newInternalError(newValidationErrorUnknownExecType(execType.string()))
	}
	return nil, nil
}

func newClientProvider(execOptions ExecOptions) (ClientProvider, error) {
	if err := validateExecOptions(execOptions); err != nil {
		return nil, err
	}
	switch execOptions.Type() {
	case ExecTypeOs:
		return newOsClientProvider(execOptions.(*OsExecOptions)), nil
	default:
		return nil, newInternalError(newValidationErrorUnknownExecType(execOptions.Type().string()))
	}
}

func validateExecOptions(execOptions ExecOptions) ValidationError {
	switch execOptions.Type() {
	case ExecTypeOs:
		return nil
	default:
		return newValidationErrorUnknownExecType(execOptions.Type().string())
	}
}
