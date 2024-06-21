package types

import (
	"errors"
	"fmt"
)

type Role string

const (
	RoleActive  Role = "active"
	RoleStandby Role = "standby"
)

var (
	errRoleIsInvalid = errors.New("role is invalid (expected: active or standby)")
)

func (r Role) Validate() error {
	if r != RoleActive && r != RoleStandby {
		return fmt.Errorf("%w: %s",
			errRoleIsInvalid, r,
		)
	}
	return nil
}
