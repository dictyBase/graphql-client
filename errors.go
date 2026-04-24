package main

import "fmt"

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
