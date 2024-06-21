package types

import (
	"encoding/json"
	"fmt"
)

type Location [36]byte

var NilLocation = Location{}

func locationSize() int {
	return 36
}

var (
	errProbeLocationStringTooLong = fmt.Errorf("probe location string too long (max %d bytes)", locationSize())
)

func NewLocation(str string) (Location, error) {
	b := []byte(str)
	if len(b) > locationSize() {
		return NilLocation, fmt.Errorf("%w: %s",
			errProbeLocationStringTooLong, str,
		)
	}

	l := Location{}
	copy(l[:], b)

	return l, nil
}

func (l *Location) String() string {
	n := 0
	for n < len(l) {
		if l[n] == 0 {
			break
		}
		n += 1
	}
	return string(l[:n])
}

func (l *Location) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	b := []byte(str)
	if len(b) > locationSize() {
		return fmt.Errorf("%w: %s",
			errProbeLocationStringTooLong, str,
		)
	}
	copy(l[:], b)
	return nil
}

func (l *Location) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	b := []byte(str)
	if len(b) > locationSize() {
		return fmt.Errorf("%w: %s",
			errProbeLocationStringTooLong, str,
		)
	}
	copy(l[:], b)
	return nil
}
