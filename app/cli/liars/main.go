package main

import (
	"fmt"
	"os"

	"github.com/ardanlabs/liarsdice/app/cli/liars/board"
	"github.com/ardanlabs/liarsdice/app/cli/liars/engine"
	"github.com/ardanlabs/liarsdice/app/cli/liars/settings"
)

const (
	keyStorePath = "zarf/ethereum/keystore/"
	passPhrase   = "123"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Parse flags for settings.

	flags, args, err := settings.Parse()
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	if _, exists := flags["h"]; exists {
		settings.PrintUsage()
		return nil
	}

	// =========================================================================
	// Establish a client connection to the game engine and get configuration.

	eng := engine.New(args.Engine)
	token, err := eng.Connect(keyStorePath, args.AccountID, passPhrase)
	if err != nil {
		return fmt.Errorf("connect to game engine: %w", err)
	}

	config, err := eng.Configuration()
	if err != nil {
		return fmt.Errorf("get game configuration: %w", err)
	}

	// =========================================================================
	// Create the board and initalize the display.

	board, err := board.New(eng, token.Address, config.Network, config.ChainID, config.ContractID)
	if err != nil {
		return err
	}
	defer board.Shutdown()

	// =========================================================================
	// Establish a websocket connection to capture the game events.

	teardown, err := eng.Events(board.Events)
	if err != nil {
		return err
	}
	defer teardown()

	// =========================================================================
	// Start handling board input.

	<-board.Run()

	return nil
}
