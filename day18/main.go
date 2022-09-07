package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"unicode"

	"github.com/k0kubun/pp"
)

type Vertex struct {
	y int
	x int
}

type Edge struct {
	v1 Vertex
	v2 Vertex
}

const (
	Entrance    rune = '@'
	Wall        rune = '#'
	OpenPassage rune = '.'
)

func findEntrance(m [][]rune) Vertex {
	for i, line := range m {
		for j, c := range line {
			if c == Entrance {
				return Vertex{i, j}
			}
		}
	}
	return Vertex{-1, -1}
}

func isMoveAllowed(s rune, doorsOpened bool, visitedKeys map[rune]bool) bool {
	if !doorsOpened {
		if s == OpenPassage ||
			s == Entrance ||
			unicode.IsLower(s) ||
			(unicode.IsUpper(s) && visitedKeys[unicode.ToLower(s)]) {
			return true
		} else {
			return false
		}
	} else if s == Wall {
		return false
	}

	return true
}

func findNeighbors(p Vertex, m [][]rune, visitedKeys map[rune]bool, doorsOpened bool) []Vertex {
	neighbors := make([]Vertex, 0)
	pl := Vertex{p.y - 1, p.x}
	if p.y-1 >= 0 && isMoveAllowed(m[pl.y][pl.x], doorsOpened, visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y + 1, p.x}
	if p.y+1 < len(m) && isMoveAllowed(m[pl.y][pl.x], doorsOpened, visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x - 1}
	if p.x-1 >= 0 && isMoveAllowed(m[pl.y][pl.x], doorsOpened, visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x + 1}
	if p.x+1 < len(m[p.y]) && isMoveAllowed(m[pl.y][pl.x], doorsOpened, visitedKeys) {
		neighbors = append(neighbors, pl)
	}

	return neighbors
}

func findIndex(Q []Vertex, p Vertex) (bool, int) {
	for i, q := range Q {
		if q == p {
			return true, i
		}
	}

	return false, -1
}

func isKeyReachable(m [][]rune, source Vertex, destination Vertex, visitedKeys map[rune]bool) bool {
	Q := make([]Vertex, 0)
	explored := make(map[Vertex]bool)
	explored[source] = true
	Q = append(Q, source)

	for len(Q) > 0 {
		v := Q[0]
		Q = Q[1:]
		if v == destination {
			return true
		}

		for _, w := range findNeighbors(v, m, visitedKeys, false) {
			if !explored[w] {
				explored[w] = true
				Q = append(Q, w)
			}
		}
	}

	return false
}

func reachableKeysBitSet(key Vertex, remainingKeysVi map[Vertex]int, remainingKeysIv map[int]Vertex, remainingKeysBitSet int, m [][]rune, visitedKeys map[rune]bool) int {
	reachable := 0

	for i := 0; remainingKeysBitSet != 0; i++ {
		if remainingKeysBitSet&1 != 0 {
			if isKeyReachable(m, key, remainingKeysIv[i], visitedKeys) {
				reachable |= 1 << i
			}
		}
		remainingKeysBitSet >>= 1
	}
	return reachable
}

type CacheKey struct {
	currentKey    int
	remainingKeys int
}

func minDistance(m [][]rune, distances map[Edge]int, currentKey Vertex, remainingKeysVi map[Vertex]int, remainingKeysIv map[int]Vertex, remainingKeysBitSet int, visitedKeys map[rune]bool, cache map[CacheKey]int) int {
	if remainingKeysBitSet == 0 {
		return 0
	}

	// make sure we take into account entrance here
	cacheKey := CacheKey{remainingKeysVi[currentKey], remainingKeysBitSet}

	if d, ok := cache[cacheKey]; ok {
		return d
	}

	min := math.MaxInt

	reachableBitset := reachableKeysBitSet(currentKey, remainingKeysVi, remainingKeysIv, remainingKeysBitSet, m, visitedKeys)
	for i := 0; reachableBitset != 0; i++ {
		if reachableBitset&1 != 0 {
			mask := 1 << i
			remainingKeysBitSet &= ^mask
			p := remainingKeysIv[i]
			visitedKeys[m[p.y][p.x]] = true
			distance := 0
			if d, ok := distances[Edge{currentKey, remainingKeysIv[i]}]; ok {
				distance = d
			} else if d, ok := distances[Edge{remainingKeysIv[i], currentKey}]; ok {
				distance = d
			} else {
				panic("invalid edge present!")
			}
			distance += minDistance(m, distances, remainingKeysIv[i], remainingKeysVi, remainingKeysIv, remainingKeysBitSet, visitedKeys, cache)
			delete(visitedKeys, m[p.y][p.x])
			remainingKeysBitSet |= mask

			if distance < min {
				min = distance
			}
		}
		reachableBitset >>= 1
	}

	cacheKey = CacheKey{remainingKeysVi[currentKey], remainingKeysBitSet}
	cache[cacheKey] = min
	return min
}

func removeKey(keys []Vertex, key Vertex) []Vertex {
	for i, k := range keys {
		if key == k {
			return append(keys[:i], keys[i+1:]...)
		}
	}

	return keys
}

func dijkstra(m [][]rune, source Vertex, destination Vertex, k map[rune]Vertex) int {
	dist := make(map[Vertex]int)
	Q := make([]Vertex, 0)

	for i := range m {
		for j := range m[i] {
			dist[Vertex{i, j}] = math.MaxInt
		}
	}

	Q = append(Q, source)

	dist[source] = 0
	visitedKeys := make(map[rune]bool)
	for len(Q) > 0 {
		u := findMinDistance(Q, dist)
		if u == destination {
			return dist[u]
		}
		if found, i := findIndex(Q, u); found {
			Q = append(Q[:i], Q[i+1:]...)
		}

		for _, v := range findNeighbors(u, m, visitedKeys, true) {
			alt := dist[u] + 1
			if alt < dist[v] {
				dist[v] = alt
				Q = append(Q, v)
			}
		}
	}
	return -1
}

func findMinDistance(Q []Vertex, dist map[Vertex]int) Vertex {
	min := math.MaxInt
	p := Vertex{-1, -1}
	for k, v := range dist {
		if found, _ := findIndex(Q, k); found && v <= min {
			min = v
			p = k
		}
	}
	return p
}

func findKeys(m [][]rune) map[rune]Vertex {
	keys := make(map[rune]Vertex)
	for i, line := range m {
		for j, s := range line {
			if unicode.IsLower(s) {
				keys[s] = Vertex{i, j}
			}
		}
	}
	return keys
}

func preCalculateDistances(m [][]rune, keysList []Vertex, keys map[rune]Vertex) map[Edge]int {
	distances := make(map[Edge]int)
	for i := 0; i < len(keysList); i++ {
		for j := 0; j < len(keysList); j++ {
			if i == j {
				continue
			}
			key := Edge{keysList[i], keysList[j]}
			if _, ok := distances[key]; ok {
				continue
			}
			key = Edge{keysList[j], keysList[i]}
			if _, ok := distances[key]; ok {
				continue
			}
			d := dijkstra(m, keysList[i], keysList[j], keys)
			distances[key] = d
		}
	}

	return distances
}

func part1(m [][]rune) int {
	entrance := findEntrance(m)
	keys := findKeys(m)
	sources := make([]rune, 0)
	for k := range keys {
		sources = append(sources, k)
	}
	// sort the sources, to have keys a, b, c, d...
	sort.Slice(sources, func(i, j int) bool {
		return sources[i] < sources[j]
	})
	sourcePoints := make([]Vertex, 0)
	// add the entrance also
	sourcePoints = append(sourcePoints, entrance)
	for _, s := range sources {
		sourcePoints = append(sourcePoints, keys[s])
	}

	remainingKeysVi := make(map[Vertex]int)
	remainingKeysIv := make(map[int]Vertex)
	remainingKeysBitSet := 0
	for i, v := range sources {
		remainingKeysVi[keys[v]] = i
		remainingKeysIv[i] = keys[v]
		remainingKeysBitSet |= 1 << i
	}

	visitedKeys := make(map[rune]bool)
	cache := make(map[CacheKey]int)
	distances := preCalculateDistances(m, sourcePoints, keys)
	return minDistance(m, distances, entrance, remainingKeysVi, remainingKeysIv, remainingKeysBitSet, visitedKeys, cache)
}

func printMap(m [][]rune) {
	for _, line := range m {
		for _, c := range line {
			fmt.Print(string(c))
		}
		fmt.Println()
	}
}

func part2(m [][]rune) int {
	entrance := findEntrance(m)
	_, _, _, _ = updateMap(entrance, m)
	keys := findKeys(m)
	sources := make([]rune, 0)
	for k := range keys {
		sources = append(sources, k)
	}
	// sort the sources, to have keys a, b, c, d...
	sort.Slice(sources, func(i, j int) bool {
		return sources[i] < sources[j]
	})
	sourcePoints := make([]Vertex, 0)
	// add the entrance also
	sourcePoints = append(sourcePoints, entrance)
	for _, s := range sources {
		sourcePoints = append(sourcePoints, keys[s])
	}

	remainingKeysVi := make(map[Vertex]int)
	remainingKeysIv := make(map[int]Vertex)
	remainingKeysBitSet := 0
	for i, v := range sources {
		remainingKeysVi[keys[v]] = i
		remainingKeysIv[i] = keys[v]
		remainingKeysBitSet |= 1 << i
	}

	visitedKeys := make(map[rune]bool)
	cache := make(map[CacheKey]int)
	distances := preCalculateDistances(m, sourcePoints, keys)
	return minDistance(m, distances, entrance, remainingKeysVi, remainingKeysIv, remainingKeysBitSet, visitedKeys, cache)
	return 0
}

func updateMap(e Vertex, m [][]rune) (Vertex, Vertex, Vertex, Vertex) {
	// centre wall
	m[e.y][e.x] = Wall
	// up wall
	m[e.y-1][e.x] = Wall
	// down wall
	m[e.y+1][e.x] = Wall
	// left wall
	m[e.y][e.x-1] = Wall
	// right wall
	m[e.y][e.x+1] = Wall
	// add the 4 entrances
	// left-up
	m[e.y-1][e.x-1] = Entrance
	// right-up
	m[e.y-1][e.x+1] = Entrance
	// left-down
	m[e.y+1][e.x-1] = Entrance
	// right-down
	m[e.y+1][e.x+1] = Entrance

	return Vertex{e.y - 1, e.x - 1}, Vertex{e.y - 1, e.x + 1}, Vertex{e.y + 1, e.x - 1}, Vertex{e.y + 1, e.x + 1}
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
	pp.Println("The result to 1st part is: ", part1(m))

	// part 2 solution
	pp.Println("The result to 2nd part is: ", part2(m))
}
