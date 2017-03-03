package utils

import "strconv"

func StringToInt64(input string) (int64, error) {
	if input == "" {
		return 0, nil
	}

	res, err := strconv.ParseInt(input, 10, 64)

	if err != nil {
		return 0, err
	}

	return res, nil
}

func StringToUint64(input string) (uint64, error) {
	if input == "" {
		return 0, nil
	}

	res, err := strconv.ParseUint(input, 10, 64)

	if err != nil {
		return 0, err
	}

	return res, nil
}

func StringToUint16(input string) (uint16, error) {
	if input == "" {
		return 0, nil
	}

	res, err := strconv.ParseUint(input, 10, 16)

	if err != nil {
		return 0, err
	}

	return uint16(res), nil
}
