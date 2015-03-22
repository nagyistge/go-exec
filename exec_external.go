package exec

type ExternalExecOptions struct {
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	TmpDir string `json:"tmp_dir,omitempty" yaml:tmp_dir,omitempty"`
}

func NewExternalExecutorReadFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorReadFileManagerProvider, error) {
	return NewExternalClientProvider(externalExecOptions)
}

func NewExternalExecutorWriteFileManagerProvider(externalExecOptions *ExternalExecOptions) (ExecutorWriteFileManagerProvider, error) {
	return NewExternalClientProvider(externalExecOptions)
}

func NewExternalClientProvider(externalExecOptions *ExternalExecOptions) (ClientProvider, error) {
	return newExternalClientProvider(externalExecOptions)
}

func ConvertExternalExecOptions(externalExecOptions *ExternalExecOptions) (ExecOptions, error) {
	return convertExternalExecOptions(externalExecOptions)
}
