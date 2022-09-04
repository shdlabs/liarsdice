// Package board handles the game board and all interactions.
package board

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ardanlabs/liarsdice/app/cli/liars/engine"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Game positioning values for static content.
const (
	boardWidth    = 63
	boardHeight   = 18
	messageHeight = 12
	anteX         = 40
	anteY         = 8
	potX          = 40
	potY          = 9
	helpX         = 65
	statusY       = 14
)

// Game positioning values for user input.
const (
	betRowX = 13
	betRowY = 10
)

// Game positioning values for changing values.
const (
	columnHeight = 2
	playersX     = 3
	outX         = 22
	betX         = 35
	balX         = 50
	myDiceX      = 13
	myDiceY      = 8
)

var words = []string{"", "one's", "two's", "three's", "four's", "five's", "six's"}

// =============================================================================

// Board represents the game board and all its state.
type Board struct {
	accountID string
	engine    *engine.Engine
	screen    tcell.Screen
	style     tcell.Style
	bets      []rune
	messages  []string
}

// New contructs a game board.
func New(engine *engine.Engine, accountID string) (*Board, error) {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("new screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}

	style := tcell.StyleDefault
	style = style.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)

	board := Board{
		accountID: accountID,
		engine:    engine,
		screen:    screen,
		style:     style,
		messages:  make([]string, 3),
	}

	return &board, nil
}

// Shutdown tearsdown the game board.
func (b *Board) Shutdown() {
	b.screen.Fini()
}

// Init generates the initial game board and starts the event loop.
func (b *Board) Init() {
	b.screen.Clear()

	for i := 1; i < boardWidth; i++ {
		b.screen.SetContent(i, 1, '=', nil, b.style)
	}
	for i := 1; i < boardWidth; i++ {
		b.screen.SetContent(i, boardHeight, '=', nil, b.style)
	}
	for i := 2; i < boardHeight; i++ {
		b.screen.SetContent(1, i, '|', nil, b.style)
	}
	for i := 2; i < boardHeight; i++ {
		b.screen.SetContent(boardWidth-1, i, '|', nil, b.style)
	}

	for i := 1; i < boardWidth; i++ {
		b.screen.SetContent(i, messageHeight, '=', nil, b.style)
	}

	b.print(3, messageHeight, " Message Center ")
	b.print(playersX, columnHeight, "Players:")
	b.print(outX, columnHeight, "Outs:")
	b.print(betX, columnHeight, "Last Bet:")
	b.print(balX, columnHeight, "  Balances:")
	b.print(myDiceX-9, myDiceY, "My Dice:")
	b.print(anteX-6, anteY, "Ante:")
	b.print(potX-6, potY, "Pot :")
	b.print(potX-6, potY+1, "Bet :")
	b.print(betRowX-9, betRowY, "My Bet :>")

	b.print(helpX, 1, "<1-6>+   : set bet")
	b.print(helpX, 2, "<del>    : remove bet number")
	b.print(helpX, 3, "<l>      : call liar")
	b.print(helpX, 4, "<n>      : new game")
	b.print(helpX, 5, "<j>      : join game")
	b.print(helpX, 6, "<s>      : start game")

	b.print(helpX, statusY-6, "status   :")
	b.print(helpX, statusY-5, "round    :")
	b.print(helpX, statusY-4, "lastbet  :")
	b.print(helpX, statusY-3, "lastwin  :")
	b.print(helpX, statusY-2, "lastlose :")
	b.print(helpX, statusY, "engine   :")
	b.print(helpX, statusY+1, "blockchn :")
	b.print(helpX, statusY+2, "chainid  :")
	b.print(helpX, statusY+3, "contract :")
	b.print(helpX, statusY+4, "account  :")

	b.bets = []rune{}
	b.screen.ShowCursor(betRowX+1, betRowY)
	b.screen.SetContent(betRowX, betRowY, ' ', nil, b.style)
	b.screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	b.print(betRowX, betRowY, "                 ")
}

// StartEventLoop starts a goroutine to handle keyboard input.
func (b *Board) StartEventLoop() chan struct{} {
	quit := make(chan struct{})

	go func() {
		for {
			ev := b.screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape:
					close(quit)
					return

				case tcell.KeyDEL:
					b.subBet()

				case tcell.KeyEnter:
					b.enterBet()

				case tcell.KeyRune:
					b.processKeyEvent(ev.Rune())
				}
			}
		}
	}()

	return quit
}

// Events handles any events from the websocket.
func (b *Board) Events(event string, address string) {
	message := fmt.Sprintf("addr: %s type: %s", b.FmtAddress(address), event)
	b.printMessage(message)

	switch event {
	case "start":
		if _, err := b.engine.RollDice(); err != nil {
			b.printMessage("error rolling dice")
		}

	case "callliar":
		b.reconcile()
	}

	status, err := b.engine.QueryStatus()
	if err != nil {
		return
	}
	b.PrintStatus(status)
}

// PrintSettings draws the settings on the board.
func (b *Board) PrintSettings(engine string, network string, chainID int, contractID string, address string) {
	b.print(helpX+11, statusY, engine)
	b.print(helpX+11, statusY+1, network)
	b.print(helpX+11, statusY+2, fmt.Sprintf("%d", chainID))
	b.print(helpX+11, statusY+3, b.FmtAddress(contractID))
	b.print(helpX+11, statusY+4, b.FmtAddress(address))
}

// PrintStatus display the status information.
func (b *Board) PrintStatus(status engine.Status) {

	// Print the current game status and round.
	b.print(helpX+11, statusY-6, fmt.Sprintf("%-10s", status.Status))
	b.print(helpX+11, statusY-5, fmt.Sprintf("%d   ", status.Round))

	// Show the account who last won and lost.
	if status.LastWinAcctID != "" {
		b.print(helpX+11, statusY-3, b.FmtAddress(status.LastWinAcctID))
		b.print(helpX+11, statusY-2, b.FmtAddress(status.LastOutAcctID))
	}

	// Show the last bet.
	if len(status.Bets) > 0 {
		bet := status.Bets[len(status.Bets)-1]
		betStr := fmt.Sprintf("%d %-10s", bet.Number, words[bet.Suite])
		b.print(helpX+11, statusY-4, betStr)
	} else {
		b.print(helpX+11, statusY-4, "                 ")
	}

	var pot float64

	// Clear the player lines.
	for i := 0; i < 5; i++ {
		addrY := columnHeight + 1 + i
		b.print(playersX, addrY, fmt.Sprintf("%-*s", boardWidth-4, " "))
		b.print(myDiceX, myDiceY, fmt.Sprintf("%-20s", " "))
	}

	for i, cup := range status.Cups {
		pot += status.AnteUSD

		// Players Column.
		addrY := columnHeight + 2 + i
		accountID := b.FmtAddress(cup.AccountID)
		b.print(playersX+3, addrY, accountID)

		// Outs.
		b.print(outX, addrY, fmt.Sprintf("%d", cup.Outs))

		// Show the active player.
		if i == status.CurrentCup {
			b.print(playersX, addrY, "->")
		} else {
			b.print(playersX, addrY, "  ")
		}

		// Last Bets.
		if cup.LastBet.Number != 0 {
			bet := fmt.Sprintf("%d %-10s", cup.LastBet.Number, words[cup.LastBet.Suite])
			b.print(betX, addrY, bet)
		} else {
			b.print(betX, addrY, "                 ")
		}

		// Balance Column.
		const balWidth = 15
		bal := fmt.Sprintf("%*s", balWidth, "$"+status.Balances[i])
		b.print(boardWidth-(balWidth+2), addrY, bal)

		// Show the dice for the connected account.
		if strings.EqualFold(cup.AccountID, b.accountID) {
			if cup.Dice[0] != 0 {
				dice := fmt.Sprintf("[%d][%d][%d][%d][%d]", cup.Dice[0], cup.Dice[1], cup.Dice[2], cup.Dice[3], status.Cups[i].Dice[4])
				b.print(myDiceX, myDiceY, dice)
			}
		}
	}

	// Show the ante and pot information.
	b.print(anteX, anteY, fmt.Sprintf("$%.2f", status.AnteUSD))
	b.print(potX, potY, fmt.Sprintf("$%.2f", pot))

	// Hide the cursor to show the game is over.
	if status.Status == "gameover" {
		b.screen.HideCursor()
	}

	b.screen.Show()
}

// FmtAddress provides a shortened version of an address.
func (*Board) FmtAddress(address string) string {
	if len(address) != 42 {
		return address
	}
	return fmt.Sprintf("%s..%s", address[:5], address[39:])
}

// =============================================================================

// processKeyEvent is the first line of processing for any key that is
// pressed during the game.
func (b *Board) processKeyEvent(r rune) {
	switch {
	case (r >= rune('0') && r <= rune('9')) || r == rune('.'):
		b.value(r)

	case r == rune('n'):
		b.newGame()

	case r == rune('j'):
		b.joinGame()

	case r == rune('s'):
		b.startGame()

	case r == rune('l'):
		b.callLiar()

	default:
		b.screen.Beep()
	}
}

// value processes the keystroke based on the mode.
func (b *Board) value(r rune) {
	if r >= rune('1') && r <= rune('6') {
		b.addBet(r)
		return
	}
	b.screen.Beep()
}

// newGame starts a new game.
func (b *Board) newGame() {
	if _, err := b.engine.NewGame(); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}
}

// joinGame adds the account to the game.
func (b *Board) joinGame() {
	status, err := b.engine.QueryStatus()
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	if status.Status != "newgame" {
		b.printMessage("error: invalid status state: " + status.Status)
		return
	}

	if _, err = b.engine.JoinGame(); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}
}

// startGame start the game so it can be played.
func (b *Board) startGame() {
	status, err := b.engine.QueryStatus()
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	if status.Status != "newgame" {
		b.printMessage("error: invalid status state: " + status.Status)
		return
	}

	if _, err := b.engine.StartGame(); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}
}

// callLiar calls the last bet a lie.
func (b *Board) callLiar() {
	status, err := b.engine.QueryStatus()
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	if status.Status != "playing" {
		b.printMessage("error: invalid status state: " + status.Status)
		return
	}

	if status.CupsOrder[status.CurrentCup] != b.accountID {
		b.printMessage("error: not your turn")
		b.screen.Beep()
		return
	}

	if _, err := b.engine.Liar(); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}
}

// reconcile the game the winner gets paid.
func (b *Board) reconcile() {
	status, err := b.engine.QueryStatus()
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	if status.Status != "gameover" {
		return
	}

	if status.LastWinAcctID != b.accountID {
		b.printMessage("gameover: winner will call reconcile")
		return
	}

	if _, err := b.engine.Reconcile(); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}
}

// =============================================================================

// addBet takes the value selected on the keyboard and adds it to the
// bet slice and screen.
func (b *Board) addBet(r rune) {
	if len(b.bets) > 0 && b.bets[0] != r {
		b.screen.Beep()
		return
	}

	x := betRowX
	b.bets = append(b.bets, r)
	x += len(b.bets)

	b.screen.ShowCursor(x+1, betRowY)
	b.print(x, betRowY, string(r))

	suite, err := strconv.Atoi(string(b.bets[0]))
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	bet := fmt.Sprintf("%d %-10s", len(b.bets), words[suite])
	b.print(potX, potY+1, bet)
}

// subBet removes a value from the bet slice and screen.
func (b *Board) subBet() {
	if len(b.bets) == 0 {
		b.screen.Beep()
		return
	}

	x := betRowX
	x += len(b.bets)
	b.bets = b.bets[:len(b.bets)-1]

	b.screen.ShowCursor(x, betRowY)
	b.print(x, betRowY, " ")

	bet := "                 "
	if len(b.bets) > 0 {
		suite, err := strconv.Atoi(string(b.bets[0]))
		if err != nil {
			b.printMessage("error: " + err.Error())
			return
		}

		bet = fmt.Sprintf("%d %-10s", len(b.bets), words[suite])
	}
	b.print(potX, potY+1, bet)
}

// enterBet is called to submit a bet.
func (b *Board) enterBet() {
	status, err := b.engine.QueryStatus()
	if err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	if status.Status != "playing" {
		b.printMessage("error: invalid status state: " + status.Status)
		return
	}

	if status.CupsOrder[status.CurrentCup] != b.accountID {
		b.printMessage("error: not your turn")
		b.screen.Beep()
		return
	}

	if len(b.bets) == 0 {
		b.printMessage("error: missing bet information")
		b.screen.Beep()
		return
	}

	if _, err = b.engine.Bet(len(b.bets), b.bets[0]); err != nil {
		b.printMessage("error: " + err.Error())
		return
	}

	b.bets = []rune{}
	b.screen.ShowCursor(betRowX+1, betRowY)
	b.print(betRowX, betRowY, "                 ")
	b.print(potX, potY+1, "                 ")
}

// =============================================================================

// PrintMessage adds a message to the message center.
func (b *Board) printMessage(message string) {
	const width = boardWidth - 4
	msg := fmt.Sprintf("%-*s", width, message)
	if len(msg) > 58 {
		msg = msg[:58]
	}

	b.messages[2] = b.messages[1]
	b.messages[1] = b.messages[0]
	b.messages[0] = msg

	b.print(3, messageHeight+2, b.messages[0])
	b.print(3, messageHeight+3, b.messages[1])
	b.print(3, messageHeight+4, b.messages[2])

	b.screen.Show()
}

// print knows how to print a string on the screen.
func (b *Board) print(x, y int, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		b.screen.SetContent(x, y, c, comb, b.style)
		x += w
	}
	b.screen.Show()
}
