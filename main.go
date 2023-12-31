package main

import (
	"github.com/gdamore/tcell"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Status string

type GameMode int64

const (
	EASY GameMode = iota
	MEDIUM
	HARD
)

const (
	height             = 25
	width              = 80
	UP                 = "UP"
	DOWN               = "DOWN"
	RIGHT              = "RIGHT"
	LEFT               = "LEFT"
	GAME_OVER   Status = "GAME OVER"
	NOT_STARTED Status = "NOT STARTED"
	STARTED     Status = "STARTED"
)

type Game struct {
	Snake
	Screen tcell.Screen
	Event  chan string
	mu     sync.Mutex
	mode   GameMode
	food   Food
	status Status
}

type Food struct {
	FoodPosition *Position
}

func newFood(x, y int) Food {
	return Food{FoodPosition: &Position{x, y}}
}
func (game *Game) printFood() {
	game.Screen.SetContent(game.food.FoodPosition.X, game.food.FoodPosition.Y, '🍔', nil, tcell.StyleDefault)
}

func (game *Game) printSnake() {
	snakeStyle := tcell.StyleDefault.Background(tcell.ColorGreen)
	for _, pos := range *game.Snake.Positions {
		game.Screen.SetContent(pos.X, pos.Y, tcell.RuneCkBoard, nil, snakeStyle)
	}
	game.printFood()

}

func main() {
	rand.Seed(time.Now().UnixNano())
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	err = screen.Init()
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite).Bold(true))

	if err != nil {
		panic(err)
	}

	game := Game{
		NewSnake(),
		screen,
		make(chan string, 10),
		sync.Mutex{},
		MEDIUM,
		newFood(rand.Intn(width), rand.Intn(height)),
		NOT_STARTED,
	}
	mode := make(chan GameMode, 1)
	go game.listenEvent(mode)
	go game.setGameMode(mode)
	defer close(game.Event)

	ticker := time.NewTicker(game.speed())
	defer ticker.Stop()
	for {
		game.Screen.Clear()
		select {
		case event := <-game.Event:
			game.Screen.Clear()
			game.UpdateSnakePosition(event)
		case <-ticker.C:
			game.updateScreen()
			ticker.Reset(game.speed())
		}
	}
}

func (game *Game) speed() time.Duration {
	//fmt.Printf("Speed \n", game.mode)
	switch game.mode {
	case EASY:
		return 800 * time.Millisecond
	case MEDIUM:
		return 500 * time.Millisecond
	case HARD:
		return 100 * time.Millisecond
	default:
		return 500 * time.Millisecond
	}
}

func (game *Game) UpdateSnakePosition(dir string) {
	existingPosition := *game.Snake.Positions
	len := len(existingPosition)
	isValid := verifyValidMovement(dir, game.Snake.Direction)
	if !isValid {
		return
	}
	game.Direction = dir
	newX := 0
	newY := 0
	if dir == UP {
		newY = -1
	} else if dir == LEFT {
		newX = -1
	} else if dir == RIGHT {
		newX = 1
	} else {
		newY = 1
	}
	newPosition := &Position{existingPosition[len-1].X + newX, existingPosition[len-1].Y + newY}
	if game.isGameOver(newPosition) {
		game.endGame()
	}
	existingPosition = append(existingPosition, *newPosition)
	if game.canEatFood(newPosition) {
		game.food = newFood(rand.Intn(width), rand.Intn(height))
	} else {
		existingPosition = existingPosition[1:]
	}
	game.Snake.Positions = &existingPosition
	game.printSnake()
}

func (game *Game) canEatFood(position *Position) bool {
	return game.food.FoodPosition.X == position.X && game.food.FoodPosition.Y == position.Y
}

func verifyValidMovement(direction string, snakeDir string) bool {
	if (direction == RIGHT && snakeDir == LEFT) || (direction == LEFT && snakeDir == RIGHT) {
		return false
	}

	if (direction == UP && snakeDir == DOWN) || (direction == DOWN && snakeDir == UP) {
		return false
	}
	return true
}

func (game *Game) listenEvent(modeChan chan GameMode) {
	for {
		event := game.Screen.PollEvent()
		switch e := event.(type) {
		case *tcell.EventResize:
			game.Screen.Sync()
		case *tcell.EventKey:
			if game.status == NOT_STARTED {
				//fmt.Printf("  \nUPDATED  %s \n", e.Name())
				if e.Key() == tcell.KeyUp && game.mode == MEDIUM {
					modeChan <- EASY
				} else if e.Key() == tcell.KeyUp && game.mode == HARD {
					modeChan <- MEDIUM
				} else if e.Key() == tcell.KeyDown && game.mode == EASY {
					modeChan <- MEDIUM
				} else if e.Key() == tcell.KeyDown && game.mode == MEDIUM {
					modeChan <- HARD
				} else if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyEnter {
					game.status = STARTED
					game.UpdateSnakePosition(game.Direction)
					close(modeChan)
					continue
				}
				continue
			}
			if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC {
				game.Screen.Fini()
				os.Exit(0)
			} else if e.Key() == tcell.KeyUp {
				game.Event <- UP
			} else if e.Key() == tcell.KeyDown {
				game.Event <- DOWN
			} else if e.Key() == tcell.KeyLeft {
				game.Event <- LEFT
			} else if e.Key() == tcell.KeyRight {
				game.Event <- RIGHT
			}
		}
	}
}

func (game *Game) updateScreen() {
	game.Screen.Clear()
	game.setBorder()
	game.loadGame()
	game.UpdateFoodAndSnakePosition()
	game.printOver()
	game.Screen.Show()
}

func (game *Game) setBorder() {
	boardStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	game.Screen.SetContent(0, 0, tcell.RuneULCorner, nil, boardStyle)
	game.Screen.SetContent(0, height, tcell.RuneLLCorner, nil, boardStyle)
	for i := 1; i < height; i++ {
		game.Screen.SetContent(0, i, tcell.RuneVLine, nil, boardStyle)
		game.Screen.SetContent(width, i, tcell.RuneVLine, nil, boardStyle)
	}
	for i := 1; i < width; i++ {
		game.Screen.SetContent(i, height, tcell.RuneHLine, nil, boardStyle)
		game.Screen.SetContent(i, 0, tcell.RuneHLine, nil, boardStyle)
	}
	game.Screen.SetContent(width, height, tcell.RuneLRCorner, nil, boardStyle)
	game.Screen.SetContent(width, 0, tcell.RuneURCorner, nil, boardStyle)
}

func (game *Game) loadGame() {
	if game.status != NOT_STARTED {
		return
	}
	x := width / 10
	y := height / 10
	displayWords := []string{"Select Game Mode : ", "EASY", "MEDIUM", "HARD", "              Press ENTER To Continue"}
	for _, word := range displayWords {
		y++
		if game.isCurrentGameMode(word) {
			word = "> " + word
		} else {
			word = "  " + word
		}
		for i, ch := range word {
			game.Screen.SetContent(x+i, y, ch, nil, tcell.StyleDefault)
		}
	}

}

func (game *Game) setGameMode(modeChan chan GameMode) {
	for {
		select {
		case x, ok := <-modeChan:
			if !ok {
				return
			} else {
				game.mu.Lock()
				game.mode = x
				game.mu.Unlock()
				game.updateScreen()

			}
		}
	}
}

func (game *Game) isCurrentGameMode(mode string) bool {
	switch game.mode {
	case EASY:
		return mode == "EASY"
	case MEDIUM:
		return mode == "MEDIUM"
	case HARD:
		return mode == "HARD"
	}
	return false
}

func (game *Game) UpdateFoodAndSnakePosition() {
	if game.status != STARTED {
		return
	}
	game.printFood()
	game.UpdateSnakePosition(game.Direction)
}

func (game *Game) isGameOver(position *Position) bool {
	if game.status == GAME_OVER {
		return true
	}
	return position != nil && (position.X == 0 || position.X == width || position.Y == height || position.Y == 0)
}

func (game *Game) endGame() {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.status = GAME_OVER
}

func (game *Game) printOver() {
	if game.isGameOver(nil) {
		for i, ch := range "GAME OVER" {
			game.Screen.SetContent(width/5+i, height/5, ch, nil, tcell.StyleDefault)
		}
	}
}
