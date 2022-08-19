// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/liarsdice/app/engine/handlers/v1/gamegrp"
	"github.com/ardanlabs/liarsdice/business/core/game"
	"github.com/ardanlabs/liarsdice/business/web/auth"
	"github.com/ardanlabs/liarsdice/foundation/events"
	"github.com/ardanlabs/liarsdice/foundation/web"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log    *zap.SugaredLogger
	Auth   *auth.Auth
	DB     *sqlx.DB
	Banker game.Banker
	Evts   *events.Events
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register group endpoints.
	ggh := gamegrp.Handlers{
		Banker: cfg.Banker,
		Evts:   cfg.Evts,
		WS:     websocket.Upgrader{},
	}

	app.Handle(http.MethodGet, version, "/game/events", ggh.Events)
	app.Handle(http.MethodGet, version, "/game/status", ggh.Status)
	app.Handle(http.MethodGet, version, "/game/new/:ante", ggh.NewGame)
	app.Handle(http.MethodGet, version, "/game/join/:address", ggh.Join)
	app.Handle(http.MethodGet, version, "/game/start", ggh.Start)
	app.Handle(http.MethodGet, version, "/game/reconcile", ggh.Reconcile)
	app.Handle(http.MethodGet, version, "/game/rolldice/:address", ggh.RollDice)
	app.Handle(http.MethodGet, version, "/game/claim/:address/:number/:suite", ggh.Claim)
	app.Handle(http.MethodGet, version, "/game/liar/:address", ggh.CallLiar)
	app.Handle(http.MethodGet, version, "/game/newround", ggh.NewRound)
	app.Handle(http.MethodGet, version, "/game/next", ggh.NextTurn)
	app.Handle(http.MethodGet, version, "/game/out/:address/:outs", ggh.UpdateOut)
	app.Handle(http.MethodGet, version, "/game/balance/:address", ggh.Balance)
}
