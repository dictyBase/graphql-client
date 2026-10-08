package main

import (
	"errors"
	"fmt"
)

// errNoOrderItems reports an order without any stock item id.
var errNoOrderItems = errors.New("at least one stock item id is required")

func errInvalidEmail(field string) error {
	return fmt.Errorf("invalid %s email: must be a valid address", field)
}

func errInvalidPlasmidType(pt PlasmidType) error {
	return fmt.Errorf(
		"invalid plasmid type: %s, must be one of: ALL, REGULAR, GOLDEN_BRAID",
		pt,
	)
}

func errInvalidStrainType(st StrainType) error {
	return fmt.Errorf(
		"invalid strain type: %s, must be one of: ALL, REGULAR, GWDI, BACTERIAL",
		st,
	)
}

func errInvalidStatus(s string) error {
	return fmt.Errorf(
		"invalid order status: %s, must be one of: IN_PREPARATION, GROWING, CANCELLED, SHIPPED",
		s,
	)
}
