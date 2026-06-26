package savedata

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type VersionTuple []int

// CompareGameVersion compares two game version strings and returns
//   - `-1` if left < right
//   - `0` if left == right
//   - `1` if left > right.
func CompareGameVersion(left string, right string) (int, error) {
	leftTuple, err := parseGameVersion(left)
	if err != nil {
		return 0, err
	}

	rightTuple, err := parseGameVersion(right)
	if err != nil {
		return 0, err
	}

	return slices.Compare(leftTuple, rightTuple), nil
}

func parseGameVersion(version string) (VersionTuple, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 || len(parts) > 4 {
		return nil, fmt.Errorf("invalid version format: %q", version)
	}

	tuple := make(VersionTuple, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid version format: %q", version)
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version component %q in %q", part, version)
		}

		tuple = append(tuple, value)
	}

	return trimTrailingZeros(tuple), nil
}

func trimTrailingZeros(tuple VersionTuple) VersionTuple {
	trimmed := len(tuple)
	for trimmed > 0 && tuple[trimmed-1] == 0 {
		trimmed--
	}

	return tuple[:trimmed]
}
