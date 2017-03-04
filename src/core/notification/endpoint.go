package notification

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type (
	Headers  map[string]string
	Endpoint struct {
		Id      uint64  `json:"id"`
		Name    string  `json:"name"`
		Url     string  `json:"url"`
		Method  string  `json:"method"`
		Headers Headers `json:"headers"`
	}
)

func (h *Headers) Value() (driver.Value, error) {
	j, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}

	return driver.Value(string(j)), nil
}

func (h *Headers) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	strValue, ok := src.(string)

	if !ok {
		return fmt.Errorf("headers field must be a string, got %T instead", src)
	}

	err := json.Unmarshal([]byte(strValue), h)

	if err != nil {
		return err
	}

	return nil
}
