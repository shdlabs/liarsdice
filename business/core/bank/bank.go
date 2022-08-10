package bank

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ardanlabs/liarsdice/contract/sol/go/contract"
	"github.com/ardanlabs/liarsdice/foundation/smartcontract/smart"
	"github.com/ethereum/go-ethereum/common"
)

// Bank represents a bank that allows for the reconciling of a game and
// information about player balances.
type Bank struct {
	client *smart.Client
}

// NewBank returns a new bank with the ability to manage the game money.
func NewBank(ctx context.Context, network string, keyPath string, passPhrase string) (*Bank, error) {
	client, err := smart.Connect(ctx, network, keyPath, passPhrase)
	if err != nil {
		return nil, err
	}

	bank := Bank{
		client: client,
	}

	return &bank, nil
}

// PlayerBalance will return the specified address balance.
func (b *Bank) PlayerBalance(ctx context.Context, address string) (*big.Int, error) {
	contract, err := newContract(b.client)
	if err != nil {
		return nil, err
	}

	tranOpts, err := b.client.NewCallOpts(ctx)
	if err != nil {
		return nil, err
	}

	player := common.HexToAddress(address)

	return contract.PlayerBalance(tranOpts, player)
}

// Reconcile will apply with ante to the winner and losers and provide the
// house the game fee.
func (b *Bank) Reconcile(ctx context.Context, winner string, losers []string, ante uint, gameFee uint) error {
	return nil
}

// newContract constructs a SimpleCoin contract.
func newContract(client *smart.Client) (*contract.Contract, error) {
	data, err := os.ReadFile("zarf/contract/id.env")
	if err != nil {
		return nil, fmt.Errorf("readfile: %w", err)
	}
	contractID := string(data)

	contract, err := contract.NewContract(common.HexToAddress(contractID), client.ContractBackend())
	if err != nil {
		return nil, fmt.Errorf("NewContract: %w", err)
	}

	return contract, nil
}