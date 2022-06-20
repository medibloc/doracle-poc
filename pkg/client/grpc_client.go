package client

import (
	"context"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"

	//"github.com/medibloc/doracle-poc/cmd/doracle-poc/mode"
	"github.com/medibloc/doracle-poc/pkg/panacea"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
	"time"

	"google.golang.org/grpc"
)

type GrpcClient struct {
	conn              *grpc.ClientConn
	interfaceRegistry codectypes.InterfaceRegistry
}

func NewGrpcClient(conf *panacea.Config, grpcAddr string) (*GrpcClient, error) {
	u, err := url.Parse(grpcAddr)
	if err != nil {
		return nil, err
	}
	prefixLen := len(u.Scheme + "://")
	addrBody := grpcAddr[prefixLen:]

	var creds credentials.TransportCredentials

	if u.Scheme == "tcp" || u.Scheme == "http" {
		creds = insecure.NewCredentials()
	} else if u.Scheme == "https" {
		creds = credentials.NewClientTLSFromCert(nil, "")
	} else {
		return nil, fmt.Errorf("invalid panacea grpc addr: %s", grpcAddr)
	}

	conn, err := grpc.Dial(addrBody, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Panacea: %w", err)
	}

	return &GrpcClient{
		conn:              conn,
		interfaceRegistry: conf.InterfaceRegistry,
	}, nil
}

func (c *GrpcClient) Broadcast(txBytes []byte) (*tx.BroadcastTxResponse, error) {
	txClient := tx.NewServiceClient(c.conn)
	return txClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
}

func (c *GrpcClient) GetAccount(panaceaAddr string) (authtypes.AccountI, error) {
	client := authtypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.Account(ctx, &authtypes.QueryAccountRequest{Address: panaceaAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to get account info via grpc: %w", err)
	}

	var acc authtypes.AccountI
	if err := c.interfaceRegistry.UnpackAny(response.GetAccount(), &acc); err != nil {
		return nil, fmt.Errorf("failed to unpack account info: %w", err)
	}
	return acc, nil
}

func (c *GrpcClient) GetRegisterOracle(address string) (*oracletypes.Oracle, error) {
	client := oracletypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Oracle(ctx, &oracletypes.QueryOracleRequest{
		Address: address,
	})

	if err != nil {
		return nil, err
	}

	return res.GetOracle(), nil
}

func (c *GrpcClient) GetOraclePublicKey() ([]byte, error) {
	client := oracletypes.NewQueryClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.OracleParams(ctx, &oracletypes.QueryOracleParamRequest{})
	if err != nil {
		return nil, err
	}

	return res.Params.OraclePublicKey, nil
}
