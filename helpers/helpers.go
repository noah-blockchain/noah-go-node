package helpers

import (
	"math/big"
)

func NoahToQNoah(noah *big.Int) *big.Int {
	p := big.NewInt(10)
	p.Exp(p, big.NewInt(18), nil)
	p.Mul(p, noah)

	return p
}
