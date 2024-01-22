// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"
	"time"

	"github.com/ardanlabs/ethereum/currency"
	"github.com/ardanlabs/liarsdice/app/services/engine/handlers/v1/gamegrp"
	"github.com/ardanlabs/liarsdice/business/core/bank"
	"github.com/ardanlabs/liarsdice/business/web/v1/auth"
	"github.com/ardanlabs/liarsdice/business/web/v1/mid"
	"github.com/ardanlabs/liarsdice/foundation/events"
	"github.com/ardanlabs/liarsdice/foundation/logger"
	"github.com/ardanlabs/liarsdice/foundation/web"
	"github.com/gorilla/websocket"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	Auth           *auth.Auth
	Converter      *currency.Converter
	Bank           *bank.Bank
	Evts           *events.Events
	AnteUSD        float64
	ActiveKID      string
	BankTimeout    time.Duration
	ConnectTimeout time.Duration
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register group endpoints.
	ggh := gamegrp.Handlers{
		Converter:      cfg.Converter,
		Bank:           cfg.Bank,
		Log:            cfg.Log,
		Evts:           cfg.Evts,
		WS:             websocket.Upgrader{},
		Auth:           cfg.Auth,
		ActiveKID:      cfg.ActiveKID,
		AnteUSD:        cfg.AnteUSD,
		BankTimeout:    cfg.BankTimeout,
		ConnectTimeout: cfg.ConnectTimeout,
	}

	app.Handle(http.MethodPost, version, "/game/connect", ggh.Connect)

	app.Handle(http.MethodGet, version, "/game/events", ggh.Events)
	app.Handle(http.MethodGet, version, "/game/config", ggh.Configuration)
	app.Handle(http.MethodGet, version, "/game/usd2wei/:usd", ggh.USD2Wei)

	app.Handle(http.MethodGet, version, "/game/status", ggh.Status, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/new", ggh.NewGame, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/join", ggh.Join, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/start", ggh.StartGame, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/rolldice", ggh.RollDice, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/bet/:number/:suite", ggh.Bet, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/liar", ggh.CallLiar, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/reconcile", ggh.Reconcile, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/balance", ggh.Balance, mid.Authenticate(cfg.Auth))

	// Timeout Situations with a player
	app.Handle(http.MethodGet, version, "/game/next", ggh.NextTurn, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/game/out/:outs", ggh.UpdateOut, mid.Authenticate(cfg.Auth))
}
