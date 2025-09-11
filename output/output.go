package output

import (
	"os"
	"encoding/json"
)

func ToJSONString(
	v interface{},
	prefix string,
	indent string,
) (string, error) {
	s, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func ToJSONFile(
	v interface{},
	filename string,
	prefix string,
	indent string,
) error {
	file, err := os.Create(filename)
	if err != nil { return err }
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent(prefix, indent)
	return encoder.Encode(v)
}