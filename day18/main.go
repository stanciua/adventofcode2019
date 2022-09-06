package main

import (
	"bufio"
	// "fmt"
	"github.com/k0kubun/pp"
	"log"
	"math"
	"os"
	"sort"
	"unicode"
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

func isAllowed(s rune, visitedKeys map[rune]bool) bool {
	if s == OpenPassage ||
		unicode.IsLower(s) ||
		(unicode.IsUpper(s) && visitedKeys[unicode.ToLower(s)]) {
		return true
	}
	return false
}

func adjacentVertices(v Vertex, edges []Edge) []Vertex {
	vertices := make([]Vertex, 0)
	for _, e := range edges {
		if e.v1 == v {
			vertices = append(vertices, e.v2)
		} else if e.v2 == v {
			vertices = append(vertices, e.v1)
		}
	}
	return vertices
}
func findNeighbors(p Vertex, m map[Vertex]rune, visitedKeys map[rune]bool) []Vertex {
	neighbors := make([]Vertex, 0)
	pl := Vertex{p.y - 1, p.x}
	if _, ok := m[pl]; ok && isAllowed(m[pl], visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y + 1, p.x}
	if _, ok := m[pl]; ok && isAllowed(m[pl], visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x - 1}
	if _, ok := m[pl]; ok && isAllowed(m[pl], visitedKeys) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x + 1}
	if _, ok := m[pl]; ok && isAllowed(m[pl], visitedKeys) {
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

func isKeyReachable(graph map[Vertex]rune, source Vertex, destination Vertex, visitedKeys map[rune]bool) bool {
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

		for _, w := range findNeighbors(v, graph, visitedKeys) {
			if !explored[w] {
				explored[w] = true
				Q = append(Q, w)
			}
		}
	}

	return false
}

func reachableKeys(key Vertex, keys []Vertex, graph map[Vertex]rune, visitedKeys map[rune]bool) []Vertex {
	reachable := make([]Vertex, 0)

	for _, v := range keys {
		if isKeyReachable(graph, key, v, visitedKeys) {
			reachable = append(reachable, v)
		}
	}
	return reachable
}

type CacheKey struct {
	key           Vertex
	remainingKeys [26]Vertex
}

func minDistance(graph map[Vertex]rune, distances map[Edge]int, currentKey Vertex, remainingKeys *[]Vertex, visitedKeys map[rune]bool, cache map[CacheKey]int) int {
	if len(*remainingKeys) == 0 {
		return 0
	}

	var remainingKeyCache [26]Vertex
	copy(remainingKeyCache[:], *remainingKeys)
	cacheKey := CacheKey{currentKey, remainingKeyCache}

	if d, ok := cache[cacheKey]; ok {
		return d
	}

	min := math.MaxInt

	reachableKeys := reachableKeys(currentKey, *remainingKeys, graph, visitedKeys)
	for _, k := range reachableKeys {
		*remainingKeys = removeKey(*remainingKeys, k)
		visitedKeys[graph[k]] = true
		distance := 0
		if d, ok := distances[Edge{currentKey, k}]; ok {
			distance = d
		} else {
			pp.Println(string(graph[currentKey]), string(graph[k]))
		}
		distance += minDistance(graph, distances, k, remainingKeys, visitedKeys, cache)
		delete(visitedKeys, graph[k])
		*remainingKeys = append(*remainingKeys, k)

		if distance < min {
			min = distance
		}
	}

	copy(remainingKeyCache[:], *remainingKeys)
	cacheKey = CacheKey{currentKey, remainingKeyCache}
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

func dijkstra(m map[Vertex]rune, source Vertex, destination Vertex, k map[rune]Vertex, doors map[rune]Vertex) int {
	dist := make(map[Vertex]int)
	prev := make(map[Vertex]Vertex)
	Q := make([]Vertex, 0)

	pkey := make(map[Vertex]rune)
	for k1, v := range k {
		pkey[v] = k1
	}
	for k := range m {
		dist[k] = math.MaxInt
		Q = append(Q, k)
	}

	dist[source] = 0
	visitedKeys := make(map[rune]bool)
	for len(Q) > 0 {
		u := findMinDistance(Q, dist)
		if u == destination {
			// get the door position relative to this key
			p := doors[unicode.ToUpper(m[u])]
			m[p] = OpenPassage
			return dist[u]
		}
		if unicode.IsLower(m[u]) {
			p := doors[unicode.ToUpper(m[u])]
			m[p] = OpenPassage
		}
		if !unicode.IsUpper(m[u]) {
			if found, i := findIndex(Q, u); found {
				Q = append(Q[:i], Q[i+1:]...)
			}
		}

		for _, v := range findNeighbors(u, m, visitedKeys) {
			if found, _ := findIndex(Q, v); found {
				alt := dist[u] + 1
				if alt < dist[v] {
					dist[v] = alt
					prev[v] = u
				}
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
func findKeysAndDoors(m [][]rune) (map[rune]Vertex, map[rune]Vertex) {
	doors := make(map[rune]Vertex)
	keys := make(map[rune]Vertex)
	for i, line := range m {
		for j, s := range line {
			if unicode.IsUpper(s) {
				doors[s] = Vertex{i, j}
				// } else if unicode.IsLower(s) || s == Entrance {
			} else if unicode.IsLower(s) {
				keys[s] = Vertex{i, j}
			}
		}
	}

	return keys, doors
}

func preCalculateDistances(graph map[Vertex]rune, keysList []Vertex, keys map[rune]Vertex, doors map[rune]Vertex) map[Edge]int {
	distances := make(map[Edge]int)
	for i := 0; i < len(keysList); i++ {
		for j := 0; j < len(keysList); j++ {
			if i == j {
				continue
			}
			d := dijkstra(graph, keysList[i], keysList[j], keys, doors)
			key := Edge{keysList[i], keysList[j]}
			if _, ok := distances[key]; ok {
				continue
			}
			distances[key] = d
		}
	}

	return distances
}

func part1(m [][]rune) int {
	entrance := findEntrance(m)
	m[entrance.y][entrance.x] = OpenPassage
	keys, doors := findKeysAndDoors(m)
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

	// build a map of all the locations except walls
	graph := make(map[Vertex]rune)
	for i := range m {
		for j := range m[i] {
			if m[i][j] != Wall {
				graph[Vertex{i, j}] = m[i][j]
			}
		}
	}

	// copy of the keys
	remainingKeys := make([]Vertex, 0)
	for _, k := range sources {
		if k == Entrance {
			continue
		}
		remainingKeys = append(remainingKeys, keys[k])
	}

	visitedKeys := make(map[rune]bool)
	copyGraph := make(map[Vertex]rune)
	for k, v := range graph {
		copyGraph[k] = v
	}
	cache := make(map[CacheKey]int)
	distances := preCalculateDistances(copyGraph, sourcePoints, keys, doors)
	return minDistance(graph, distances, entrance, &remainingKeys, visitedKeys, cache)
}

func part2(m [][]rune) int {
	return 0
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
