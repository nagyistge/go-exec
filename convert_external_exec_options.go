package exec

func convertExternalExecOptions(externalExecOptions *ExternalExecOptions) (ExecOptions, error) {
	execType, err := ExecTypeOf(externalExecOptions.Type)
	if err != nil {
		return nil, err
	}
	switch execType {
	case ExecTypeOs:
		return &OsExecOptions{
			TmpDir: externalExecOptions.TmpDir,
		}, nil
	default:
		return nil, UnknownExecType(execType)
	}
	return nil, nil
}
