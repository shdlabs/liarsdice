// Package gamegrp provides the handlers for game play.
package gamegrp

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	v1Web "github.com/ardanlabs/liarsdice/business/web/v1"
	"github.com/gorilla/websocket"

	"github.com/ardanlabs/liarsdice/business/core/game"
	"github.com/ardanlabs/liarsdice/foundation/events"
	"github.com/ardanlabs/liarsdice/foundation/signature"
	"github.com/ardanlabs/liarsdice/foundation/web"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	Banker game.Banker
	WS     websocket.Upgrader
	Evts   *events.Events

	game *game.Game
	mu   sync.RWMutex
}

// SetGame safely sets a game pointer.
func (h *Handlers) setGame(game *game.Game) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.game = game
}

// GetGame safely returns a copy of the game pointer.
func (h *Handlers) getGame() (*game.Game, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.game == nil {
		return nil, v1Web.NewRequestError(errors.New("no game exists"), http.StatusBadRequest)
	}

	return h.game, nil
}

// Events handles a web socket to provide events to a client.
func (h *Handlers) Events(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return v1Web.NewRequestError(errors.New("web value missing from context"), http.StatusBadRequest)
	}

	// Need this to handle CORS on the websocket.
	h.WS.CheckOrigin = func(r *http.Request) bool { return true }

	// This upgrades the HTTP connection to a websocket connection.
	c, err := h.WS.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// This provides a channel for receiving events from the blockchain.
	ch := h.Evts.Acquire(v.TraceID)
	defer h.Evts.Release(v.TraceID)

	// Starting a ticker to send a ping message over the websocket.
	ticker := time.NewTicker(time.Second)

	// Block waiting for events from the blockchain or ticker.
	for {
		select {
		case msg, wd := <-ch:

			// If the channel is closed, release the websocket.
			if !wd {
				return nil
			}

			if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return err
			}

		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return nil
			}
		}
	}
}

// Status will return information about the game.
func (h *Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	status := g.Info()

	var cups []Cup
	for _, cup := range status.Cups {
		cups = append(cups, Cup{Account: cup.Account, Outs: cup.Outs})
	}

	var claims []Claim
	for _, claim := range status.Claims {
		claims = append(claims, Claim{Account: claim.Account, Number: claim.Number, Suite: claim.Suite})
	}

	resp := struct {
		Status        string   `json:"status"`
		LastOutAcct   string   `json:"last_out"`
		LastWinAcct   string   `json:"last_win"`
		CurrentPlayer string   `json:"current_player"`
		CurrentCup    int      `json:"current_cup"`
		Round         int      `json:"round"`
		Cups          []Cup    `json:"cups"`
		CupsOrder     []string `json:"player_order"`
		Claims        []Claim  `json:"claims"`
	}{
		Status:        status.Status,
		LastOutAcct:   status.LastOutAcct,
		LastWinAcct:   status.LastWinAcct,
		CurrentPlayer: status.CurrentPlayer,
		CurrentCup:    status.CurrentCup,
		Round:         status.Round,
		Cups:          cups,
		CupsOrder:     status.CupsOrder,
		Claims:        claims,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// NewGame creates a new game if there is no game or the status of the current game
// is GameOver.
func (h *Handlers) NewGame(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if g, err := h.getGame(); err == nil {
		status := g.Info()
		if status.Status != game.StatusGameOver {
			return v1Web.NewRequestError(errors.New("game is currently being played"), http.StatusBadRequest)
		}
	}

	ante, err := strconv.ParseInt(web.Param(r, "ante"), 10, 64)
	if err != nil {
		return v1Web.NewRequestError(errors.New("invalid ante value"), http.StatusBadRequest)
	}

	h.setGame(game.New(h.Banker, ante))

	h.Evts.Send("newgame")

	return h.Status(ctx, w, r)
}

// Join adds the given player to the game.
func (h *Handlers) Join(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	if err := g.AddAccount(ctx, address); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("join:" + address)

	return h.Status(ctx, w, r)
}

// Start creates a new game if there is no game or the status of the current game
// is GameOver.
func (h *Handlers) Start(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	status := g.Info()
	if status.Status != game.StatusGameOver {
		return v1Web.NewRequestError(errors.New("current game status doesn't allow this call"), http.StatusBadRequest)
	}

	if err := g.StartPlay(); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("start")

	return h.Status(ctx, w, r)
}

// RollDice will roll 5 dice for the given player and game.
func (h *Handlers) RollDice(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	if err := g.RollDice(address); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	status := g.Info()
	cup, exists := status.Cups[address]
	if !exists {
		return v1Web.NewRequestError(errors.New("address not found"), http.StatusBadRequest)
	}

	h.Evts.Send("rolldice:" + address)

	resp := struct {
		Dice []int `json:"dice"`
	}{
		Dice: cup.Dice,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Claim processes a claim made by a player in a game.
func (h *Handlers) Claim(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	number, err := strconv.Atoi(web.Param(r, "number"))
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("converting number: %s", err), http.StatusBadRequest)
	}

	suite, err := strconv.Atoi(web.Param(r, "suite"))
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("converting suite: %s", err), http.StatusBadRequest)
	}

	if err := g.Claim(address, number, suite); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("claim")

	return h.Status(ctx, w, r)
}

// CallLiar processes the claims and defines a winner and a loser for the round.
func (h *Handlers) CallLiar(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	winner, loser, err := g.CallLiar(address)
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("callliar")

	resp := struct {
		Winner string `json:"winner"`
		Loser  string `json:"loser"`
	}{
		Winner: winner,
		Loser:  loser,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// NewRound starts a new round reseting the required data.
func (h *Handlers) NewRound(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	playersLeft, err := g.NextRound()
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("newround")

	resp := struct {
		PlayersLeft int `json:"players_left"`
	}{
		PlayersLeft: playersLeft,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// NextTurn changes the account that will make the next move.
func (h *Handlers) NextTurn(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	g.NextTurn()

	h.Evts.Send("nextturn")

	return h.Status(ctx, w, r)
}

// UpdateOut replaces the current out amount of the player. This call is not
// part of the game flow, it is used to control when a player should be removed
// from the game.
func (h *Handlers) UpdateOut(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	outs, err := strconv.Atoi(web.Param(r, "outs"))
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("converting outs: %s", err), http.StatusBadRequest)
	}

	if err := g.ApplyOut(address, outs); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	h.Evts.Send("outs:" + address)

	return h.Status(ctx, w, r)
}

// Balance returns the player balance from the smart contract.
func (h *Handlers) Balance(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	address := web.Param(r, "address")

	balance, err := g.PlayerBalance(ctx, address)
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusInternalServerError)
	}

	resp := struct {
		Balance *big.Int `json:"balance"`
	}{
		Balance: balance,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Reconcile calls the smart contract reconcile method.
func (h *Handlers) Reconcile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	g, err := h.getGame()
	if err != nil {
		return err
	}

	err = g.Reconcile(ctx)
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusInternalServerError)
	}

	return h.Status(ctx, w, r)
}

func (h *Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var dt struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Signature string `json:"sig"`
	}

	if err := web.Decode(r, &dt); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	fmt.Println("*********************************************")
	fmt.Printf("%#v\n", dt)
	fmt.Println("*********************************************")

	sig, err := hex.DecodeString(dt.Signature[2:])
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	fmt.Println("SIGL:", len(sig), "SIGB:", sig)
	fmt.Println("*********************************************")

	data := struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}{
		Name:   dt.Name,
		Status: dt.Status,
	}

	address, err := signature.FromAddress(data, dt.Signature)
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	fmt.Println("ADDR:", address)
	fmt.Println("*********************************************")

	return web.Respond(ctx, w, dt, http.StatusOK)
}
