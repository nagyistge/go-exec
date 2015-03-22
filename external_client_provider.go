package exec

func newExternalClientProvider(externalExecOptions *ExternalExecOptions) (ClientProvider, error) {
	execOptions, err := ConvertExternalExecOptions(externalExecOptions)
	if err != nil {
		return nil, err
	}
	return newClientProvider(execOptions)
}
