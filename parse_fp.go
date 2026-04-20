package main

import (
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	R "github.com/IBM/fp-go/v2/record"
)

func ParsePlasmidType(s string) E.Either[error, PlasmidType] {
	return F.Pipe2(
		plasmidTypeMap,
		R.Lookup[PlasmidType](s),
		E.FromOption[PlasmidType](func() error {
			return errInvalidPlasmidType(PlasmidType(s))
		}),
	)
}
