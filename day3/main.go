package main

import (
	"bufio"
	"fmt"
	"log"
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

	displayGrid(grid)
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(grid))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(grid))
}

func part1(grid [][]rune) int {
	return -1
}

func part2(grid [][]rune) int {
	return -1
}

type direction int

const (
	up direction = iota
	down
	left
	right
)

func directinFromString(s string) direction {
	var d direction

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
func buildInitialGrid(initialGridSize int) [][] rune {
	grid := make([][]rune, initialGridSize * 2)
	for idx := range grid {
		grid[idx] = make([]rune, initialGridSize * 2)
		for innerIdx := range grid[idx] {
			grid[idx][innerIdx] = '.'
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

func getGridOrigin(grid [][]rune) (int, int) {
	return len(grid) / 2, len(grid[0]) / 2
}

func resizeGrid(grid [][]rune) [][]rune {

	newGrid := make([][]rune, len(grid) * 2)
	for idx := range newGrid {
		newGrid[idx] = make([]rune, len(grid[0]) * 2)
		for innerIdx := range newGrid[idx] {
			grid[idx][innerIdx] = '.'
		}
	}
	
// 1 2
// 3 4

// 0 0 0 0
// 0 1 2 0
// 0 3 4 0
// 0 0 0 0
}
