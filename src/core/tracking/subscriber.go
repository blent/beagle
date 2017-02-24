package tracking

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
)

type (
	Headers    map[string]string
	Data       map[string]string
	Subscriber struct {
		gorm.Model
		Name    string  `json:"name" gorm:"unique_index"`
		Event   string  `json:"event" gorm:"index"`
		Method  string  `json:"method"`
		Url     string  `json:"url"`
		Headers Headers `json:"headers" gorm:"type:varchar(255)"`
		Data    Data    `json:"data" gorm:"type:varchar(255)"`
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
