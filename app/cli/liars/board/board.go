// Package board handles the game board and all interactions.
package board

import (
	"fmt"
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
	statusY       = 13
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
	betX         = 26
	balX         = 50
	myDiceX      = 13
	myDiceY      = 8
)

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
	b.print(betX, columnHeight, "Last Bet:")
	b.print(balX, columnHeight, "  Balances:")
	b.print(myDiceX-9, myDiceY, "My Dice:")
	b.print(anteX-6, anteY, "Ante:")
	b.print(potX-6, potY, "Pot :")
	b.print(betRowX-9, betRowY, "My Bet :>")
	b.print(helpX, 2, "<1-6>    : set/increment bet")
	b.print(helpX, 3, "<delete> : decrement bet")
	b.print(helpX, 4, "<s>      : start game")
	b.print(helpX, 5, "<l>      : call liar")
	b.print(helpX, statusY-6, "status   :")
	b.print(helpX, statusY-5, "round    :")
	b.print(helpX, statusY-4, "lastwin  :")
	b.print(helpX, statusY-3, "lastlose :")
	b.print(helpX, statusY, "engine   :")
	b.print(helpX, statusY+1, "blockchn :")
	b.print(helpX, statusY+2, "chainid  :")
	b.print(helpX, statusY+3, "contract :")
	b.print(helpX, statusY+4, "account  :")

	b.bets = []rune{}
	b.screen.ShowCursor(betRowX+1, betRowY)
	b.screen.SetContent(betRowX, betRowY, ' ', nil, b.style)
	b.screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	b.print(betRowX, betRowY, "                             ")
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
	case "join":
		status, err := b.engine.QueryStatus()
		if err != nil {
			return
		}
		b.PrintStatus(status)

	case "start":
		if _, err := b.engine.RollDice(); err != nil {
			b.printMessage("error rolling dice")
			return
		}
		status, err := b.engine.QueryStatus()
		if err != nil {
			return
		}
		b.PrintStatus(status)
	}
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
	b.print(helpX+11, statusY-6, status.Status)
	b.print(helpX+11, statusY-5, fmt.Sprintf("%d", status.Round))

	// Show the account who last won and lost.
	if status.LastWinAcctID != "" {
		b.print(helpX+11, statusY-4, b.FmtAddress(status.LastWinAcctID))
		b.print(helpX+11, statusY-3, b.FmtAddress(status.LastOutAcctID))
	}

	var pot float64

	for i, cup := range status.Cups {
		pot += status.AnteUSD

		// Players Column.
		addrY := columnHeight + 2 + i
		accountID := b.FmtAddress(cup.AccountID)
		b.print(playersX+3, addrY, accountID)
		b.print(betX, addrY, "")

		// Show the active player.
		if i == status.CurrentPlayer {
			b.print(playersX, addrY, "->")
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

	case r == rune('s'):
		b.startGame()

	case r == rune('l'):
		b.printMessage("CALL LIAR")

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

// startGame puts the game in playing mode.
func (b *Board) startGame() {
	status, err := b.engine.StartGame()
	if err != nil {
		b.printMessage("error: " + err.Error())
	}
	b.PrintStatus(status)
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
}

// enterBet is called to submit a bet.
func (b *Board) enterBet() {
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
