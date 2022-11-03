package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
)

const (
	Bug        rune = '#'
	EmptySpace rune = '.'
	Subgrid    rune = '?'
)

type Grid = [5][5]rune

type Tile struct {
	x     int
	y     int
	level int
}

func neighborsLevelTile(grids map[int]Grid, tile Tile) []Tile {
	neighbors := make([]Tile, 0)
	if tile.y-1 >= 0 {
		neighbors = append(neighbors, Tile{tile.x, tile.y - 1, tile.level})
	}
	if tile.y+1 < 5 {
		neighbors = append(neighbors, Tile{tile.x, tile.y + 1, tile.level})
	}
	if tile.x-1 >= 0 {
		neighbors = append(neighbors, Tile{tile.x - 1, tile.y, tile.level})
	}
	if tile.x+1 < 5 {
		neighbors = append(neighbors, Tile{tile.x + 1, tile.y, tile.level})
	}

	if _, ok := grids[tile.level-1]; ok {
		// now check the outer level for first and last line/column
		if tile.y == 0 {
			neighbors = append(neighbors, Tile{2, 1, tile.level - 1})
		}
		if tile.x == 0 {
			neighbors = append(neighbors, Tile{1, 2, tile.level - 1})
		}
		if tile.y == 4 {
			neighbors = append(neighbors, Tile{2, 3, tile.level - 1})
		}
		if tile.x == 4 {
			neighbors = append(neighbors, Tile{3, 2, tile.level - 1})
		}
	}

	if _, ok := grids[tile.level+1]; ok {
		// now check the inner level for (2,1), (3,2), (2, 3), (1,2)
		if tile.x == 2 && tile.y == 1 {
			neighbors = append(neighbors, Tile{0, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{1, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{2, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{3, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 0, tile.level + 1})
		}
		if tile.x == 3 && tile.y == 2 {
			neighbors = append(neighbors, Tile{4, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 1, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 2, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 3, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 4, tile.level + 1})
		}
		if tile.x == 2 && tile.y == 3 {
			neighbors = append(neighbors, Tile{0, 4, tile.level + 1})
			neighbors = append(neighbors, Tile{1, 4, tile.level + 1})
			neighbors = append(neighbors, Tile{2, 4, tile.level + 1})
			neighbors = append(neighbors, Tile{3, 4, tile.level + 1})
			neighbors = append(neighbors, Tile{4, 4, tile.level + 1})
		}
		if tile.x == 1 && tile.y == 2 {
			neighbors = append(neighbors, Tile{0, 0, tile.level + 1})
			neighbors = append(neighbors, Tile{0, 1, tile.level + 1})
			neighbors = append(neighbors, Tile{0, 2, tile.level + 1})
			neighbors = append(neighbors, Tile{0, 3, tile.level + 1})
			neighbors = append(neighbors, Tile{0, 4, tile.level + 1})
		}
	}

	return neighbors
}

func countBugsInNeighbors(grids map[int]Grid, neighbors []Tile) int {
	count := 0
	for _, t := range neighbors {
		if grids[t.level][t.y][t.x] == Bug {
			count++
		}
	}

	return count
}

func computeNewGrid(grids map[int]Grid, level int) (bool, Grid) {
	infestation := false

	grid := grids[level]
	newGrid := grids[level]

	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[i]); j++ {
			if len(grids) > 1 && i == 2 && j == 2 {
				continue
			}
			neighbors := neighborsLevelTile(grids, Tile{j, i, level})
			count := countBugsInNeighbors(grids, neighbors)
			if grid[i][j] == Bug {
				// count is 1, the bug lives
				if count == 1 {
					newGrid[i][j] = grid[i][j]
					infestation = true
				} else {
					newGrid[i][j] = EmptySpace
				}
			} else {
				// count is 1 or 2, the empty space becomes  a bug
				if count == 1 || count == 2 {
					newGrid[i][j] = Bug
					infestation = true
				} else {
					newGrid[i][j] = grid[i][j]
				}
			}
		}
	}

	return infestation, newGrid
}

func simulateGrids(grids map[int]Grid) map[int]Grid {
	nextGrids := make(map[int]Grid)
	infestation := false
	replaceGrids(nextGrids, grids)

	// level 0, 1, 2, ...
	for level := 0; level <= len(grids); level++ {
		var newGrid Grid
		ok := false
		if _, ok = grids[level]; !ok {
			if infestation {
				nextGrids[level] = emptyGrid()
			}
			break
		}

		infestation, newGrid = computeNewGrid(grids, level)
		nextGrids[level] = newGrid
	}

	infestation = false
	// level -1, -2, -3 ...
	for level := -1; level >= -len(grids); level-- {
		var newGrid Grid
		ok := false
		if _, ok = grids[level]; !ok {
			if infestation {
				nextGrids[level] = emptyGrid()
			}
			break
		}

		infestation, newGrid = computeNewGrid(grids, level)
		nextGrids[level] = newGrid
	}

	return nextGrids
}

func computeBiodiversityRating(grid Grid) int {
	rating := 0
	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[i]); j++ {
			if grid[i][j] == Bug {
				rating += int(math.Pow(2, float64(i*5+j)))
			}
		}
	}

	return rating
}

func part1(m [][]rune) int {
	var grid Grid

	for i := 0; i < len(m); i++ {
		copy(grid[i][:], m[i])
	}

	matchLayout := make(map[Grid]int)
	matchLayout[grid] = 1
	rating := 0
	grids := make(map[int]Grid)
	grids[0] = grid

	for {
		_, newArea := computeNewGrid(grids, 0)
		if _, ok := matchLayout[newArea]; !ok {
			matchLayout[newArea] = 1
		} else {
			// found the same layout again, we are done
			rating = computeBiodiversityRating(newArea)
			break
		}
		grids[0] = newArea
	}

	return rating
}

func part2(m [][]rune) int {
	var grid Grid

	for i := 0; i < len(m); i++ {
		copy(grid[i][:], m[i])
	}

	grid[2][2] = Subgrid

	grids := make(map[int]Grid)
	grids[0] = grid
	grids[-1] = emptyGrid()
	grids[1] = emptyGrid()

	for i := 0; i < 200; i++ {
		newGrids := simulateGrids(grids)
		replaceGrids(grids, newGrids)
	}

	count := 0
	for _, grid := range grids {
		for i := 0; i < len(grid); i++ {
			for j := 0; j < len(grid[i]); j++ {
				if grid[i][j] == Bug {
					count++
				}
			}
		}
	}
	return count
}

func replaceGrids(curr map[int]Grid, next map[int]Grid) {
	for k, v := range next {
		curr[k] = v
	}
}

func emptyGrid() Grid {
	var grid Grid
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			grid[i][j] = EmptySpace
		}
	}

	grid[2][2] = Subgrid

	return grid
}

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

	m := make([][]rune, 0)
	for _, line := range inputs {
		r := make([]rune, 0)
		for _, s := range line {
			r = append(r, s)
		}
		m = append(m, r)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(m))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(m))
}
