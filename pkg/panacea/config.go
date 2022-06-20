package panacea

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

type Config struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Marshaller        *codec.ProtoCodec
}

func NewConfig() *Config {
	interfaceRegistry := makeInterfaceRegistry()
	return &Config{
		InterfaceRegistry: interfaceRegistry,
		Marshaller:        codec.NewProtoCodec(interfaceRegistry),
	}
}

func makeInterfaceRegistry() codectypes.InterfaceRegistry {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	oracletypes.RegisterInterfaces(interfaceRegistry)
	return interfaceRegistry
}
