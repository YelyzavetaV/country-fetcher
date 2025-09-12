package output

import (
	"os"
	"encoding/json"
	"fmt"
)

func toJSONBytes(
	v interface{},
	prefix string,
	indent string,
) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func toJSONFile(
	v interface{},
	filename string,
	prefix string,
	indent string,
	perm os.FileMode,
) error {
	b, err := toJSONBytes(v, prefix, indent)
	if err != nil {
		return err
	}

	fileExisted := false
	if _, err := os.Stat(filename); err == nil {
		fileExisted = true
	}

	if err = os.WriteFile(filename, b, perm); err != nil {
		return err
	}

	// Force desired permissions if file already existed with different
	// permissions
	if fileExisted { return os.Chmod(filename, perm) }
	return nil
}

func ToJSON(
	v interface{},
	filename string,
	prefix string,
	indent string,
	perm os.FileMode,
) error {
	if filename != "" {
		return toJSONFile(v, filename, prefix, indent, perm)
	}

	b, err := toJSONBytes(v, prefix, indent)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}