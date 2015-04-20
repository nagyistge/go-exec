package exec

import (
	"bufio"
	"io"
	"strings"
)

func readLines(readFileManager ReadFileManager, path string) (retValue []string, retErr error) {
	file, err := readFileManager.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	reader := bufio.NewReader(file)
	lines := make([]string, 0)
	for line, err := reader.ReadString('\n'); err != io.EOF; line, err = reader.ReadString('\n') {
		if err != nil {
			return nil, err
		}
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) > 0 {
			lines = append(lines, trimmedLine)
		}
	}
	return lines, nil
}
