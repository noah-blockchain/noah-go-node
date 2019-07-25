package events

import "github.com/tendermint/go-amino"

func RegisterAminoEvents(codec *amino.Codec) {
	codec.RegisterInterface((*Event)(nil), nil)
	codec.RegisterConcrete(RewardEvent{},
		"noax/RewardEvent", nil)
	codec.RegisterConcrete(SlashEvent{},
		"noax/SlashEvent", nil)
	codec.RegisterConcrete(UnbondEvent{},
		"noax/UnbondEvent", nil)
	codec.RegisterConcrete(CoinLiquidationEvent{},
		"noax/CoinLiquidationEvent", nil)
}

type Role byte

func (r Role) String() string {
	switch r {
	case RoleValidator:
		return "Validator"
	case RoleDelegator:
		return "Delegator"
	case RoleDAO:
		return "DAO"
	case RoleDevelopers:
		return "Developers"
	}

	return "Undefined"
}

const (
	RoleValidator Role = iota
	RoleDelegator
	RoleDAO
	RoleDevelopers
)

type Event interface{}
type Events []Event
