package utils

import (
	"github.com/pkg/errors"
	"reflect"
)

func SetStructField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return errors.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return errors.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	valueType := val.Type()

	if structFieldType != valueType {
		if !valueType.ConvertibleTo(structFieldType) {
			return errors.Errorf(
				"Provided value type %s didn't match obj field %s type %s",
				valueType.Name(),
				name,
				structFieldType.Name(),
			)
		}

		converted := val.Convert(structFieldType)

		structFieldValue.Set(converted)

		return nil
	}

	structFieldValue.Set(val)

	return nil
}

func MapToStruct(target interface{}, values map[string]interface{}) error {
	for key, value := range values {
		err := SetStructField(target, key, value)

		if err != nil {
			return err
		}
	}

	return nil
}
