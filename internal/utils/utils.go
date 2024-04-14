package utils

import (
	"fmt"
	"strconv"
)

func ParsePositiveInt(s string) (int, error) {
	if s == "" {
		return -1, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return -1, err
	}
	return i, nil
}

func MakeCacheKey(featureID, tagID int) string {
	return fmt.Sprintf("feature%d-tag%d", featureID, tagID)
}
