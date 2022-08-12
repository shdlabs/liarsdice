package gamegrp

// Game exposes the required Game data for the HTTP responses.
type Game struct {
	Status        string   `json:"status"`
	Round         int      `json:"round"`
	CurrentPlayer string   `json:"current_player"`
	CupsOrder     []string `json:"player_order"`
	Players       []Player `json:"players"`
}

// Player exposes the required Player data for the HTTP responses.
type Player struct {
	Wallet string `json:"wallet"`
	Outs   uint8  `json:"outs"`
	Dice   []int  `json:"dice"`
	Claim  Claim  `json:"claim"`
}

// Claim exposes the require Claim data for the HTTP requests and responses.
type Claim struct {
	Wallet string `json:"wallet"`
	Number int    `json:"number"`
	Suite  int    `json:"suite"`
}
