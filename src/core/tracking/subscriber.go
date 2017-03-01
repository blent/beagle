package tracking

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type (
	Headers    map[string]string
	Data       map[string]string
	Subscriber struct {
		Id      int64   `json:"id"`
		Name    string  `json:"name"`
		Event   string  `json:"event"`
		Method  string  `json:"method"`
		Url     string  `json:"url"`
		Enabled bool    `json:"enabled"`
		Headers Headers `json:"headers"`
		Data    Data    `json:"data"`
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

func (h *Data) Value() (driver.Value, error) {
	j, err := json.Marshal(h)

	if err != nil {
		return nil, err
	}

	return driver.Value(string(j)), nil
}

func (h *Data) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	strValue, ok := src.(string)

	if !ok {
		return fmt.Errorf("data field must be a string, got %T instead", src)
	}

	err := json.Unmarshal([]byte(strValue), h)

	if err != nil {
		return err
	}

	return nil
}
