package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"unicode"
)

type Tile struct {
	y int
	x int
}

type Portal struct {
	e1       Tile
	e2       Tile
	distance int
}

const (
	Wall        rune = '#'
	OpenPassage rune = '.'
	EmptySpace  rune = ' '
)

func findPortals(m [][]rune) map[string]Portal {
	t := Tile{0, 0}
	portals := make(map[string]Portal)
	var sb strings.Builder
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[i]); j++ {
			if m[i][j] == OpenPassage {
				// check each neighbor for upper case letter
				// up -> down -> left -> right
				if unicode.IsUpper(m[i-1][j]) && unicode.IsUpper(m[i-2][j]) {
					t = Tile{i, j}
					sb.WriteRune(m[i-2][j])
					sb.WriteRune(m[i-1][j])
				} else if unicode.IsUpper(m[i+1][j]) && unicode.IsUpper(m[i+2][j]) {
					t = Tile{i, j}
					sb.WriteRune(m[i+1][j])
					sb.WriteRune(m[i+2][j])
				} else if unicode.IsUpper(m[i][j-1]) && unicode.IsUpper(m[i][j-2]) {
					t = Tile{i, j}
					sb.WriteRune(m[i][j-2])
					sb.WriteRune(m[i][j-1])
				} else if unicode.IsUpper(m[i][j+1]) && unicode.IsUpper(m[i][j+2]) {
					t = Tile{i, j}
					sb.WriteRune(m[i][j+1])
					sb.WriteRune(m[i][j+2])
				} else {
					continue
				}

				name := sb.String()
				t0 := Tile{0, 0}
				if p, ok := portals[name]; ok {
					if p.e1 == t0 {
						portals[name] = Portal{t, p.e2, 0}
					} else {
						portals[name] = Portal{p.e1, t, 0}
					}
				} else {
					portals[name] = Portal{Tile{0, 0}, t, 0}
				}

				sb.Reset()
			}
		}
	}
	return portals
}

func splitPortals(m [][]rune, portals map[string]Portal) (map[Tile]string, map[Tile]string) {
	outerPortals, innerPortals := make(map[Tile]string), make(map[Tile]string)
	for i := 3; i < len(m)-3; i++ {
		for j := 3; j < len(m[i])-3; j++ {
			if m[i][j] == OpenPassage {
				t := Tile{i, j}
				for k, v := range portals {
					if v.e1 == t || v.e2 == t {
						innerPortals[t] = k
					}
				}
			}
		}
	}

	for k, v := range portals {
		if k == "AA" || k == "ZZ" {
			continue
		}
		if _, ok := innerPortals[v.e1]; ok {
			outerPortals[v.e2] = k
		} else {
			outerPortals[v.e1] = k
		}
	}

	return innerPortals, outerPortals
}

func connections(portals map[string]Portal) map[Tile]Tile {
	conns := make(map[Tile]Tile)
	for k, c := range portals {
		if k == "AA" || k == "ZZ" {
			continue
		}

		conns[c.e1] = c.e2
		conns[c.e2] = c.e1
	}

	return conns
}

func modifiedDijkstra(level int, m [][]rune, source Tile, destination Tile, innerPortals map[Tile]string, outerPortals map[Tile]string, conns map[Tile]Tile, portals map[string]Portal) int {
	dist := make(map[Tile]int)
	Q := make([]Tile, 0)

	for i := range m {
		for j := range m[i] {
			dist[Tile{i, j}] = math.MaxInt
		}
	}

	Q = append(Q, source)

	dist[source] = 0
	for len(Q) > 0 {
		u, idx := findMinDistance(Q, dist)
		if u == destination {
			return dist[u]
		}

		Q = append(Q[:idx], Q[idx+1:]...)

		for _, v := range modifiedFindNeighbors(level, u, m, portals, innerPortals, outerPortals, conns) {
			step := 1
			if _, ok := conns[u]; ok {
				step++
			}
			alt := dist[u] + step
			if alt < dist[v] {
				dist[v] = alt
				Q = append(Q, v)
			}
		}
	}

	return -1
}

func dijkstra(m [][]rune, source Tile, destination Tile, portals map[string]Portal, conns map[Tile]Tile) int {
	dist := make(map[Tile]int)
	Q := make([]Tile, 0)

	for i := range m {
		for j := range m[i] {
			dist[Tile{i, j}] = math.MaxInt
		}
	}

	Q = append(Q, source)

	dist[source] = 0
	for len(Q) > 0 {
		u, idx := findMinDistance(Q, dist)
		if u == destination {
			return dist[u]
		}

		Q = append(Q[:idx], Q[idx+1:]...)

		for _, v := range findNeighbors(u, m, portals, conns) {
			step := 1
			if _, ok := conns[u]; ok {
				step++
			}
			alt := dist[u] + step
			if alt < dist[v] {
				dist[v] = alt
				Q = append(Q, v)
			}
		}
	}

	return -1
}

func findMinDistance(Q []Tile, dist map[Tile]int) (Tile, int) {
	min := math.MaxInt
	p := Tile{-1, -1}
	idx := 0
	for k, v := range dist {
		if found, i := findIndex(Q, k); found && v <= min {
			min = v
			p = k
			idx = i
		}
	}
	return p, idx
}

func findIndex(Q []Tile, p Tile) (bool, int) {
	for i, q := range Q {
		if q == p {
			return true, i
		}
	}

	return false, -1
}

func modifiedFindNeighbors(level int, p Tile, m [][]rune, portals map[string]Portal, innerPortals map[Tile]string, outerPortals map[Tile]string, conns map[Tile]Tile) []Tile {
	neighbors := make([]Tile, 0)

	for _, t := range []Tile{{p.y - 1, p.x}, {p.y + 1, p.x}, {p.y, p.x - 1}, {p.y, p.x + 1}} {
		if isAllowedModified(level, t, m, portals, innerPortals, outerPortals, conns) {
			neighbors = append(neighbors, t)
		}
	}
	return neighbors
}

func findNeighbors(p Tile, m [][]rune, portals map[string]Portal, conns map[Tile]Tile) []Tile {
	neighbors := make([]Tile, 0)

	for _, t := range []Tile{{p.y - 1, p.x}, {p.y + 1, p.x}, {p.y, p.x - 1}, {p.y, p.x + 1}} {
		if isAllowed(t, m, portals, conns) {
			// if this is a portal add the other connection
			if c, ok := conns[t]; ok {
				neighbors = append(neighbors, c)
			} else {
				neighbors = append(neighbors, t)
			}
		}
	}
	return neighbors
}

func isAllowed(t Tile, m [][]rune, portals map[string]Portal, conns map[Tile]Tile) bool {
	if !(t.y >= 0 && t.y < len(m) && t.x >= 0 && t.x < len(m[0])) {
		return false
	}

	if _, ok := conns[t]; !ok && m[t.y][t.x] != OpenPassage {
		return false
	}

	return true
}

func isAllowedModified(level int, t Tile, m [][]rune, portals map[string]Portal, innerPortals map[Tile]string, outerPortals map[Tile]string, conns map[Tile]Tile) bool {
	if !(t.y >= 0 && t.y < len(m) && t.x >= 0 && t.x < len(m[0])) {
		return false
	}

	if !isPortalEnabled(level, t, innerPortals, outerPortals) {
		return false
	}

	if _, ok := conns[t]; !ok && m[t.y][t.x] != OpenPassage {
		return false
	}

	return true
}

func isPortalEnabled(level int, t Tile, innerPortals map[Tile]string, outerPortals map[Tile]string) bool {
	// if this is not a portal you can pass
	if _, ok := innerPortals[t]; !ok {
		if _, ok := outerPortals[t]; !ok {
			return true
		}
	}

	if level == 0 {
		if outerPortals[t] == "AA" || outerPortals[t] == "ZZ" {
			return true
		}
		if _, ok := innerPortals[t]; ok {
			return true
		}
		return false
	} else if level > 0 {
		if outerPortals[t] == "AA" || outerPortals[t] == "ZZ" {
			return false
		}
	}
	return true
}

// 26, 396
func minDistance(level int, source Tile, m [][]rune, innerPortals map[Tile]string, outerPortals map[Tile]string, conns map[Tile]Tile, portals map[string]Portal) int {
	if p, ok := innerPortals[source]; ok {
		fmt.Println(p)
	} else if p, ok := outerPortals[source]; ok {
		fmt.Println(p)
	}
	if level > 5 || outerPortals[source] == "ZZ" && isPortalEnabled(level, source, innerPortals, outerPortals) {
		return 0
	}

	min := math.MaxInt

	reachablePortals := reachablePortals(level, source, m, innerPortals, outerPortals, conns, portals)
	for _, p := range reachablePortals {
		distance := p.distance
		// if source and destination are the same check to see where are we going up or down 1 level
		if _, ok := outerPortals[p.e2]; ok {
			level--
		} else if _, ok := innerPortals[p.e2]; ok {
			level++
		}
		// when we jmp to another level we do 1 step
		distance += 1

		distance += minDistance(level, p.e2, m, innerPortals, outerPortals, conns, portals)

		if distance < min {
			min = distance
		}
	}
	return min
}

func reachablePortals(level int, source Tile, m [][]rune, innerPortals map[Tile]string, outerPortals map[Tile]string, conns map[Tile]Tile, portals map[string]Portal) []Portal {
	reachable := make([]Portal, 0)
	if level == 0 {
		for p := range innerPortals {
			if d := modifiedDijkstra(level, m, source, p, innerPortals, outerPortals, conns, portals); d != -1 {
				reachable = append(reachable, Portal{source, p, d})
			}
		}

		// also check for the destination ZZ
		dest := portals["ZZ"].e2
		if d := modifiedDijkstra(level, m, source, dest, innerPortals, outerPortals, conns, portals); d != -1 {
			reachable = append(reachable, Portal{source, dest, d})
		}
	} else if level > 0 {
		for p := range innerPortals {
			if d := modifiedDijkstra(level, m, source, p, innerPortals, outerPortals, conns, portals); d != -1 {
				reachable = append(reachable, Portal{source, p, d})
			}
		}
		for p, n := range outerPortals {
			if n == "AA" || n == "ZZ" {
				continue
			}
			if d := modifiedDijkstra(level, m, source, p, innerPortals, outerPortals, conns, portals); d != -1 {
				reachable = append(reachable, Portal{source, p, d})
			}
		}
	}

	return reachable
}

func part1(m [][]rune) int {
	portals := findPortals(m)
	conns := connections(portals)
	source := portals["AA"].e2
	destination := portals["ZZ"].e2
	return dijkstra(m, source, destination, portals, conns)
}

func part2(m [][]rune) int {
	portals := findPortals(m)
	innerPortals, outerPortals := splitPortals(m, portals)
	conns := connections(portals)
	source := portals["AA"].e2
	return minDistance(0, source, m, innerPortals, outerPortals, conns, portals)
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
