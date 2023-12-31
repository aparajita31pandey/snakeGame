package main

type Snake struct {
	Positions *[]Position
	Direction string
}

func NewSnake() Snake {
	initialPositions := []Position{{X: 1, Y: 2}, {2, 3}, {3, 4}, {4, 5}} // Create a slice of Position and initialize it with a Position instance
	return Snake{&initialPositions, RIGHT}
}

type Position struct {
	X int
	Y int
}
