package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
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

	orbitsStrMap := make(map[string]([]string))
	strToIntMap := make(map[string](int))
	idx := 0
	for _, line := range inputs {
		orbit := strings.Split(line, ")")
		if _, ok := strToIntMap[orbit[0]]; !ok {
			strToIntMap[orbit[0]] = idx
			idx++
		}
		if _, ok := strToIntMap[orbit[1]]; !ok {
			strToIntMap[orbit[1]] = idx
			idx++
		}
		if objects, present := orbitsStrMap[orbit[0]]; present {
			orbitsStrMap[orbit[0]] = append(objects, orbit[1])
		} else {
			orbitsStrMap[orbit[0]] = []string{orbit[1]}
		}
	}

	// convert string objects to ints in order to use graph vertices as indexes
	orbitsIntMap := convertToOrbitIntMap(orbitsStrMap, strToIntMap)

	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(orbitsIntMap, strToIntMap))

	// find source "YOU" and destiation "SAN" objects they are orbiting
	source, destination := findSourceAndDestinationObject(strToIntMap["YOU"], strToIntMap["SAN"], orbitsIntMap)
	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(source, destination, orbitsIntMap, strToIntMap))
}

func contains(e int, list []int) bool {
	for _, v := range list {
		if v == e {
			return true
		}
	}
	return false
}

func findSourceAndDestinationObject(source int, destination int, orbitsIntMap map[int]([]int)) (int, int) {
	s, d := 0, 0
	for k, v := range orbitsIntMap {
		if contains(source, v) {
			s = k
		} else if contains(destination, v) {
			d = k
		}
	}

	return s, d
}
func convertToOrbitIntMap(orbitsMap map[string]([]string), strToIntMap map[string](int)) map[int]([]int) {
	orbitsIntMap := make(map[int]([]int))
	for key, val := range orbitsMap {
		var newList []int
		for _, e := range val {
			newList = append(newList, strToIntMap[e])
		}

		orbitsIntMap[strToIntMap[key]] = newList
	}
	return orbitsIntMap
}

func part1(orbitsIntMap map[int]([]int), strToIntMap map[string](int)) int {
	distances := dijkstra(0, orbitsIntMap, strToIntMap)
	sum := 0
	for _, v := range distances {
		sum += v
	}

	return sum
}

func part2(source, destination int, orbitsIntMap map[int]([]int), strToIntMap map[string](int)) int {
	distances := dijkstra(source, orbitsIntMap, strToIntMap)
	return distances[destination]
}

func minDistance(vertices, dist []int) int {
	min := math.MaxInt32
	v := 0
	for _, vertex := range vertices {
		if dist[vertex] < min {
			min = dist[vertex]
			v = vertex
		}
	}

	return v
}

func remove(e int, list []int) []int {
	for i, v := range list {
		if v == e {
			return append(list[:i], list[i+1:]...)
		}
	}

	return list
}

func convertToTwoWayGraph(orbitsIntMap map[int]([]int)) map[int]([]int) {
	twoWayGraph := make(map[int]([]int))
	// copy the current one way graph
	for k, v := range orbitsIntMap {
		twoWayGraph[k] = v
	}

	for k, v := range orbitsIntMap {
		for _, e := range v {
			if val, ok := twoWayGraph[e]; ok {
				twoWayGraph[e] = append(val, k)
			} else {
				twoWayGraph[e] = []int{k}
			}
		}
	}

	return twoWayGraph
}

func dijkstra(source int, orbitsIntMap map[int]([]int),
	strToIntMap map[string](int)) []int {
	var q []int
	dist := make([]int, len(strToIntMap))
	dist[source] = 0
	twoWayGraph := convertToTwoWayGraph(orbitsIntMap)
	for _, v := range strToIntMap {
		if v != source {
			dist[v] = math.MaxInt32
		}
		q = append(q, v)
	}

	for len(q) > 0 {
		v := minDistance(q, dist)
		q = remove(v, q)

		for _, u := range twoWayGraph[v] {
			alt := dist[v] + 1
			if alt < dist[u] {
				dist[u] = alt
			}
		}
	}

	return dist
}
