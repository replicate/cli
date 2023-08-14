package util

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
)

func ParseInputs(ctx context.Context, args []string, stdin string, sep string) (map[string]string, error) {
	re := regexp.MustCompile(`{{(.*?)}}`)

	inputs := make(map[string]string)
	for _, e := range args {
		k, v, found := strings.Cut(e, sep)
		if !found {
			return nil, fmt.Errorf("invalid input: %s", e)
		}

		var stdinJSON map[string]interface{}
		if stdin != "" {
			err := json.Unmarshal([]byte(stdin), &stdinJSON)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal stdin: %w", err)
			}
		}

		// Extract data from JSON
		matches := re.FindAllStringSubmatch(v, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			path := strings.TrimSpace(match[1])
			if !strings.HasPrefix(path, "$") {
				path = "$" + path
			}

			value, err := jsonpath.Get(path, stdinJSON)
			if err != nil {
				return nil, fmt.Errorf("failed to extract data from JSON using path '%s': %w", path, err)
			}

			// Replace the segment with the extracted value
			v = strings.Replace(v, match[0], fmt.Sprintf("%v", value), 1)
		}

		// Read from file
		if strings.HasPrefix(v, "@") {
			path := strings.TrimSpace(v[1:])
			downloadURL, err := UploadFile(ctx, path)
			if err != nil {
				return nil, fmt.Errorf("failed to upload file: %w", err)
			}

			v = downloadURL
		}

		inputs[k] = v
	}

	return inputs, nil
}

func GetPipedArgs() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if info.Mode()&os.ModeNamedPipe != 0 {
		reader := bufio.NewReader(os.Stdin)
		var output []rune

		for {
			input, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				break
			}
			output = append(output, input)
		}

		return string(output), nil
	} else {
		return "", nil
	}
}
