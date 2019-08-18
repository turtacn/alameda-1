package utils

import (
	"fmt"
	"bufio"
	"os"
)

func CreateFile(filename string) error {
	if !FileExists(filename) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Errorf(err.Error())
			return err
		}
		defer file.Close()
	} else {
		fmt.Errorf("file(%s) already exists", filename)
	}
	return nil
}

func WriteFile(filename string, lines []string) error {
	// Open file using READ & WRITE permission
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	defer file.Close()

	// Write text line-by-line to file
	for _, line := range lines {
		if _, err = file.WriteString(line); err != nil {
			fmt.Errorf(err.Error())
			return err
		}
	}

	// Save changes
	err = file.Sync()
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	return nil
}

func WriteFileWithAppend(filename string, lines []string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	defer file.Close()

	// Write text line-by-line to file
	for _, line := range lines {
		if _, err = file.WriteString(line); err != nil {
			fmt.Errorf(err.Error())
			return err
		}
	}

	return nil
}

func DeleteFile(filename string) error {
	if FileExists(filename) {
		err := os.Remove(filename)
		if err != nil {
			fmt.Errorf(err.Error())
			return err
		}
	}
	return nil
}

func ReadFile(filename string) ([]string, error) {
	content := make([]string, 0)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Errorf(err.Error())
		return content, err
	}
	defer file.Close()

	// Read file, line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Errorf(err.Error())
		return make([]string, 0), err
	}

	return content, nil
}

func FileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
