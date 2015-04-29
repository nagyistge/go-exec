package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func readLines(readFileManager ReadFileManager, path string) (retValue []string, retErr error) {
	if err := checkFileExists(readFileManager, path); err != nil {
		return nil, err
	}
	file, err := readFileManager.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return getLines(file)
}

func getLines(reader io.Reader) ([]string, error) {
	bufReader := bufio.NewReader(reader)
	lines := make([]string, 0)
	for line, err := bufReader.ReadString('\n'); err != io.EOF; line, err = bufReader.ReadString('\n') {
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

func readAll(readFileManager ReadFileManager, path string) (retValue []byte, retErr error) {
	if err := checkFileExists(readFileManager, path); err != nil {
		return nil, err
	}
	file, err := readFileManager.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return ioutil.ReadAll(file)
}

func checkFileExists(readFileManager ReadFileManager, path string) error {
	exists, err := readFileManager.IsFileExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("exec: file does not exist: %s", path)
	}
	return nil
}
