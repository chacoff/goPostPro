package main

import(
	"errors"
	"strconv"
)


//getCageNumber gets the cage number to help in the LTC logic
//lint:ignore U1000 Ignore unused function temporarily for debugging
func getCageNumber(s string) (int, error) {
	if len(s) == 0 {
		return 0, errors.New("input string is empty")
	}

	lastChar := s[len(s)-1]

	digit, err := strconv.Atoi(string(lastChar))
	if err != nil {
		return 0, errors.New("failed to convert character to digit")
	}

	return digit, nil
}
