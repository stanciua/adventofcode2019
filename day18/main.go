package main

import (
	"bufio"
	"fmt"
	// pp "github.com/k0kubun/pp/v3"
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

func isAllowed(s rune) bool {
	if s == OpenPassage ||
		unicode.IsLower(s) {
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
func findNeighbors(p Vertex, m map[Vertex]rune) []Vertex {
	neighbors := make([]Vertex, 0)
	pl := Vertex{p.y - 1, p.x}
	if _, ok := m[pl]; ok && isAllowed(m[pl]) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y + 1, p.x}
	if _, ok := m[pl]; ok && isAllowed(m[pl]) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x - 1}
	if _, ok := m[pl]; ok && isAllowed(m[pl]) {
		neighbors = append(neighbors, pl)
	}
	pl = Vertex{p.y, p.x + 1}
	if _, ok := m[pl]; ok && isAllowed(m[pl]) {
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

// procedure DFS_iterative(G, v) is
//
//	let S be a stack
//	S.push(v)
//	while S is not empty do
//	    v = S.pop()
//	    if v is not labeled as discovered then
//	        label v as discovered
//	        for all edges from v to w in G.adjacentEdges(v) do
//	            S.push(w)
// func dfs(vertices []Vertex, edges []Edge, graph map[Edge]int, source Vertex, keys map[rune]Vertex) int {
// 	min := math.MaxInt
// 	// paths := make([][]string, 0)
//
// 	S := make([]Vertex, 0)
// 	discovered := make(map[Vertex]bool)
// 	S = append(S, source)
// 	for len(S) > 0 {
// 		v := S[len(S)-1]
// 		S = S[:len(S)-1]
//
// 		fmt.Println(S)
// 		if isValidPath(S, vertices) {
// 			fmt.Println(S)
// 			if steps := steps(S, graph); steps < min {
// 				min = steps
// 			}
// 		}
//
// 		if !discovered[v] {
// 			discovered[v] = true
// 			for _, w := range adjacentVertices(v, edges) {
// 				S = append(S, w)
// 			}
// 		}
// 	}
//
// 	fmt.Println(graph)
// 	return min
// }

// func steps(path []Vertex, graph map[Edge]int) int {
// 	sum := 0
// 	for i := 0; i < len(path)-1; i++ {
// 		s := graph[Edge{path[i], path[i+1]}] + graph[Edge{path[i+1], path[i]}]
// 		fmt.Println(s)
// 		sum += s
// 	}
// 	return sum
// }
//
// func isValidPath(path []Vertex, sourcePoints []Vertex) bool {
// 	// if len(path) != len(sourcePoints) {
// 	// 	return false
// 	// }
// 	ms := make(map[Vertex]bool)
// 	for _, v := range sourcePoints {
// 		ms[v] = true
// 	}
//
// 	mp := make(map[Vertex]bool)
// 	for _, v := range path {
// 		mp[v] = true
// 	}
//
// 	for k := range ms {
// 		if !mp[k] {
// 			return false
// 		}
// 	}
// 	return true
// }

func dijkstraRedo(vertices []Vertex, edges []Edge, graph map[Edge]int, source Vertex, keys map[rune]Vertex) [][]string {
	dist := make(map[Vertex]int)
	prev := make(map[Vertex]Vertex)
	Q := make([]Vertex, 0)
	// pp.Println(graph)
	for _, k := range vertices {
		dist[k] = math.MaxInt
		Q = append(Q, k)
	}

	dist[source] = 0
	for len(Q) > 0 {
		u := findMinDistance(Q, dist)
		if found, i := findIndex(Q, u); found {
			Q = append(Q[:i], Q[i+1:]...)
		}

		for _, v := range adjacentVertices(u, edges) {
			if found, _ := findIndex(Q, v); found {
				steps := 0
				if s1, ok := graph[Edge{u, v}]; ok {
					steps = s1
				} else if s2, ok := graph[Edge{v, u}]; ok {
					steps = s2
				}
				alt := dist[u] + steps
				if alt < dist[v] {
					dist[v] = alt
					prev[v] = u
				}
			}
		}
	}

	pkey := make(map[Vertex]rune)
	for k1, v := range keys {
		pkey[v] = k1
	}
	// reconstruct all the paths
	paths := make([][]string, 0)
	for _, u := range vertices[1:] {
		S := make([]string, 0)
		if _, ok := prev[u]; ok || u == source {
			for _, ok := prev[u]; ok; u, ok = prev[u] {
				S = append([]string{string(pkey[u])}, S...)
			}
		}
		paths = append(paths, S)
	}

	// pp.Println(paths)
	return paths
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

		for _, v := range findNeighbors(u, m) {
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
			} else if unicode.IsLower(s) || s == Entrance {
				keys[s] = Vertex{i, j}
			}
		}
	}

	return keys, doors
}

func part1(m [][]rune) int {
	entrance := findEntrance(m)
	keys, doors := findKeysAndDoors(m)
	m[entrance.y][entrance.x] = OpenPassage
	sources := make([]rune, 0)
	for k := range keys {
		sources = append(sources, k)
	}
	// sort the sources, to have keys a, b, c, d...
	sort.Slice(sources, func(i, j int) bool {
		return sources[i] < sources[j]
	})
	sourcePoints := make([]Vertex, 0)
	sourcePoints = append(sourcePoints, entrance)
	for _, s := range sources {
		sourcePoints = append(sourcePoints, keys[s])
	}

	// build a map of all the locations except walls
	minMap := make(map[Vertex]rune)
	for i := range m {
		for j := range m[i] {
			if m[i][j] != Wall {
				minMap[Vertex{i, j}] = m[i][j]
			}
		}
	}

	graph := make(map[Edge]int)
	edges := make([]Edge, 0)
	for _, s1 := range sourcePoints {
		for _, s2 := range sourcePoints {
			if s1 == s2 {
				continue
			}
			e := Edge{s1, s2}
			ei := Edge{s2, s1}
			_, ok1 := graph[e]
			_, ok2 := graph[ei]
			if ok1 || ok2 {
				continue
			}
			edges = append(edges, Edge{s1, s2})
			step := dijkstra(minMap, s1, s2, keys, doors)
			graph[e] = step
		}
	}

	pkey := make(map[Vertex]rune)
	for k1, v := range keys {
		pkey[v] = k1
	}
	for _, e := range edges {
		fmt.Println(string(pkey[e.v1]), " -> ", string(pkey[e.v2]))
	}
	// for _, s := range sources {
	// 	paths := dijkstraRedo(sourcePoints, edges, graph, keys[s], keys)
	// 	for _, p := range paths {
	// 		pp.Println(p)
	// 		pp.Println("----------------------------------------")
	// 	}
	// }

	return -1
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
	fmt.Println("The result to 1st part is: ", part1(m))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(m))
}
