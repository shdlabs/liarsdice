package gamegrp

import (
	"github.com/ardanlabs/liarsdice/business/core/game"
	"github.com/ethereum/go-ethereum/common"
)

type appState struct {
	GameID          string           `json:"gameID"`
	Status          string           `json:"status"`
	AnteUSD         float64          `json:"anteUSD"`
	PlayerLastOut   common.Address   `json:"lastOut"`
	PlayerLastWin   common.Address   `json:"lastWin"`
	PlayerTurn      common.Address   `json:"currentID"`
	Round           int              `json:"round"`
	Cups            []appCup         `json:"cups"`
	ExistingPlayers []common.Address `json:"playerOrder"`
	Bets            []appBet         `json:"bets"`
	Balances        []string         `json:"balances"`
}

func toAppState(state game.State, anteUSD float64, cups []appCup, bets []appBet) appState {
	return appState{
		GameID:          state.GameID,
		Status:          state.Status,
		AnteUSD:         anteUSD,
		PlayerLastOut:   state.PlayerLastOut,
		PlayerLastWin:   state.PlayerLastWin,
		PlayerTurn:      state.PlayerTurn,
		Round:           state.Round,
		Cups:            cups,
		ExistingPlayers: state.ExistingPlayers,
		Bets:            bets,
		Balances:        state.Balances,
	}
}

type appBet struct {
	Player common.Address `json:"account"`
	Number int            `json:"number"`
	Suit   int            `json:"suit"`
}

func toAppBet(bet game.Bet) appBet {
	return appBet{
		Player: bet.Player,
		Number: bet.Number,
		Suit:   bet.Suit,
	}
}

type appCup struct {
	Player  common.Address `json:"account"`
	Dice    []int          `json:"dice"`
	LastBet appBet         `json:"lastBet"`
	Outs    int            `json:"outs"`
}

func toAppCup(cup game.Cup, dice []int) appCup {
	return appCup{
		Player: cup.Player,
		Dice:   dice,
		Outs:   cup.Outs,
	}
}
