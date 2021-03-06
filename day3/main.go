package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("input/part1.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var inputs []string
	for scanner.Scan() {
		inputs = append(inputs, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// we have just two lines, with elements separated by commas
	if len(inputs) != 2 {
		panic("The input should be only two lines long!")
	}

	wirePath1 := getWirePath(inputs[0])
	wirePath2 := getWirePath(inputs[1])

	initialGridSize := getMaximumGridSize(wirePath1, wirePath2)

	// build the initial grid
	grid := buildInitialGrid(initialGridSize)

	var stepsWire1, stepsWire2 []Coordinate
	// plot the path for wire 1
	grid, stepsWire1 = plotPath(grid, wirePath1)
	// plot the path for wire 2
	grid, stepsWire2 = plotPath(grid, wirePath2)
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(grid))
	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(grid, stepsWire1, stepsWire2))
}

func part1(grid [][]rune) int {
	var intersections []Coordinate
	for i, _ := range grid {
		for j, e := range grid[i] {
			if e == 'X' {
				intersections = append(intersections, Coordinate{x: i, y: j})
			}
		}
	}

	return getManhattanDistanceToClosestPoint(getGridOrigin(grid), intersections)
}

func part2(grid [][]rune, stepsWire1, stepsWire2 []Coordinate) int {
	noSteps1 := 0
	noSteps2 := 0
	seenIntersections1 := make(map[Coordinate]int)
	seenIntersections2 := make(map[Coordinate]int)

	for _, e := range stepsWire1 {
		noSteps1++
		if grid[e.x][e.y] == 'X' {
			seenIntersections1[e] = noSteps1
		}
	}
	for _, e := range stepsWire2 {
		noSteps2++
		if grid[e.x][e.y] == 'X' {
			seenIntersections2[e] = noSteps2
		}
	}

	// now enumerate each wire intersection and check the minimum number of steps
	// for the sum of the two wires path
	min := int(^uint(0) >> 1)
	for k, v := range seenIntersections1 {
		if v2, present := seenIntersections2[k]; present && (v+v2 < min) {
			min = v + v2
		}
	}

	return min
}

type Direction int

const (
	up Direction = iota
	down
	left
	right
)

type Coordinate struct {
	x int
	y int
}

func manhattanDistance(p1 Coordinate, p2 Coordinate) int {
	return int(math.Abs(float64(p1.x-p2.x)) + math.Abs(float64(p1.y-p2.y)))
}

func getManhattanDistanceToClosestPoint(origin Coordinate, intersections []Coordinate) int {
	min := int(^uint(0) >> 1)
	for _, intersection := range intersections {
		distance := manhattanDistance(origin, intersection)
		if distance < min {
			min = distance
		}
	}

	return min
}

func directionFromString(s string) Direction {
	var d Direction

	switch s {
	case "U":
		d = up
	case "D":
		d = down
	case "L":
		d = left
	case "R":
		d = right
	default:
		panic("Invalid direction received!")
	}

	return d
}

// get the wire path from input
func getWirePath(input string) []string {
	var path []string
	for _, move := range strings.Split(input, ",") {
		path = append(path, move)
	}

	return path
}

// find the maximum bounds of the grid
func getMaximumGridSize(path1 []string, path2 []string) int {
	// combine both paths into a new path
	var combinedPath []string
	combinedPath = append(combinedPath, path1...)
	combinedPath = append(combinedPath, path2...)
	var moves []int
	for _, move := range combinedPath {
		if moveStep, err := strconv.Atoi(move[1:]); err != nil {
			panic(err)
		} else {
			moves = append(moves, moveStep)
		}
	}

	// now we need to get the maximum number of steps a move can take
	sort.Ints(moves)
	return moves[len(moves)-1]
}

// build initial grid starting from the maxumum size calculated from the input
func buildInitialGrid(initialGridSize int) [][]rune {
	grid := make([][]rune, initialGridSize*2)
	for idx := range grid {
		grid[idx] = make([]rune, initialGridSize*2)
		for innerIdx := range grid[idx] {
			grid[idx][innerIdx] = ' '
		}
	}

	return grid
}

func displayGrid(grid [][]rune) {
	for i := range grid {
		for _, r := range grid[i] {
			fmt.Print(string(r))
		}
		fmt.Println()
	}
}

func getGridOrigin(grid [][]rune) Coordinate {
	return Coordinate{x: len(grid) / 2, y: len(grid[0]) / 2}
}

func resizeGrid(grid [][]rune) [][]rune {
	newGrid := make([][]rune, len(grid)*2)
	for idx := range newGrid {
		newGrid[idx] = make([]rune, len(grid[0])*2)
		for innerIdx := range newGrid[idx] {
			newGrid[idx][innerIdx] = ' '
		}
	}

	startRowPos := (len(newGrid) - len(grid)) / 2
	startColPos := (len(newGrid[0]) - len(grid[0])) / 2
	for i := startRowPos; i < startRowPos+len(grid); i++ {
		for j := startColPos; j < startColPos+len(grid[0]); j++ {
			newGrid[i][j] = grid[i-startRowPos][j-startColPos]
		}
	}

	return newGrid
}

type void struct{}

var member void

// this function is responsible of plotting the path of each our onto the grid,
// and also for returning the trail of each wire path throug the grid
func plotPath(grid [][]rune, path []string) ([][]rune, []Coordinate) {
	// stores the steps taken by this wire, we will use this later when finding
	// out which intersection is reached by lowest number of steps
	var positions []Coordinate
	// stores the steps taken by this wire, we use this for searching if a square
	// of the grid has been seen or not
	seenPositions := make(map[Coordinate]void)

	origin := getGridOrigin(grid)
	// get the direction for the first step
	direction := directionFromString(path[0][:1])
	// we need to know the last step in order to know when to use '+'
	lastStep := getSymbolForDirection(direction)
	grid[origin.x][origin.y] = 'o'
	curr := origin
	for _, step := range path {
		direction := directionFromString(step[:1])
		symbol := getSymbolForDirection(direction)
		if lastStep != symbol {
			grid[curr.x][curr.y] = '+'
		}
		offset := Coordinate{x: 0, y: 0}
		switch direction {
		case up:
			offset.x = -1
		case down:
			offset.x = 1
		case left:
			offset.y = -1
		case right:
			offset.y = 1
		}

		noOfSteps, err := strconv.Atoi(step[1:])

		if err != nil {
			panic(err)
		}

		for i := 0; i < noOfSteps; i++ {
			// we need to check if we can move in that direction, if not
			// resize the grid
			if curr.x+offset.x >= len(grid) ||
				curr.x+offset.x < 0 ||
				curr.y+offset.y >= len(grid[0]) ||
				curr.y+offset.y < 0 {
				originOffset := Coordinate{x: curr.x - origin.x, y: curr.y - origin.y}
				grid = resizeGrid(grid)
				newOrigin := getGridOrigin(grid)
				positions = updatePositionsWithNewOrigin(positions, origin, newOrigin)
				for k := range seenPositions {
					delete(seenPositions, k)
				}
				for _, e := range positions {
					seenPositions[e] = member
				}
				curr = Coordinate{x: newOrigin.x + originOffset.x, y: newOrigin.y + originOffset.y}
				curr = Coordinate{x: newOrigin.x + originOffset.x, y: newOrigin.y + originOffset.y}
				// the new origin will be set
				origin = newOrigin
			}

			curr = Coordinate{curr.x + offset.x, curr.y + offset.y}
			_, present := seenPositions[curr]
			seenPositions[curr] = member
			positions = append(positions, curr)
			currSymbol := grid[curr.x][curr.y]
			if currSymbol != symbol && currSymbol != ' ' && !present {
				grid[curr.x][curr.y] = 'X'
			} else {
				grid[curr.x][curr.y] = symbol
			}
			lastStep = symbol
		}
	}

	return grid, positions
}

func updatePositionsWithNewOrigin(positions []Coordinate, oldOrigin Coordinate, newOrigin Coordinate) []Coordinate {
	var newPositions []Coordinate
	for _, v := range positions {
		newPositions = append(newPositions, Coordinate{x: newOrigin.x + (v.x - oldOrigin.x), y: newOrigin.y + (v.y - oldOrigin.y)})
	}

	return newPositions
}

func getSymbolForDirection(direction Direction) rune {
	symbol := 'x'
	switch direction {
	case up:
		symbol = '|'
	case down:
		symbol = '|'
	case left:
		symbol = '-'
	case right:
		symbol = '-'
	default:
		panic("Invalid direction received!")
	}

	return symbol
}
