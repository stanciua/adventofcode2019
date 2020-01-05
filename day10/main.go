package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

const epsilon = 1e-9

type Quadrant int

const (
	One   Quadrant = 1
	Four           = 4
	Three          = 3
	Two            = 2
)

type Point struct {
	x int
	y int
}

type SameSlope struct {
	points []Point
	slope  float64
}

type Position int

const (
	Middle Position = iota
	Left
	Right
)

func getListOfAsteroids(grid [][]int) []Point {
	var points []Point
	for i, _ := range grid {
		for j, _ := range grid[i] {
			if grid[i][j] == 1 {
				point := Point{x: i, y: j}
				points = append(points, point)
			}
		}
	}

	return points
}

func cmp(a, b float64) int {
	if a-b < epsilon && math.Abs(a-b) > epsilon {
		return -1
	} else if a-b > epsilon && math.Abs(a-b) > epsilon {
		return 1
	}
	return 0
}

func getSlope(p1, p2 Point) float64 {
	return float64(p2.y-p1.y) / float64(p2.x-p1.x)
}

func distance(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow(float64(p2.x-p1.x), 2) + math.Pow(float64(p2.y-p1.y), 2))
}

func isBetween(p1, p2, p3 Point) bool {
	return cmp(distance(p1, p2)+distance(p2, p3), distance(p1, p3)) == 0
}

func findListForSlope(slope float64, sameSlopeList []SameSlope) (*SameSlope, bool) {
	var sameSlope *SameSlope
	for i, v := range sameSlopeList {
		if cmp(v.slope, slope) == 0 {
			sameSlope = &sameSlopeList[i]
			return sameSlope, true
		}
	}

	return sameSlope, false
}

func getAsteroidsInRageForStation(station Point, asteroids []Point) []Point {
	var asteroidsInRange []Point
	var sameSlopeList []SameSlope
	p1 := station
	for _, p2 := range asteroids {
		if p2.x-p1.x == 0 {
			// convention for same line asteroids:
			// slope : max value of float64
			slope := math.MaxFloat64
			if sameSlope, ok := findListForSlope(slope, sameSlopeList); ok {
				sameSlope.points = append(sameSlope.points, p2)
			} else {
				sameSlopeList = append(sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
			}
		} else if p2.y-p1.y == 0 {
			// convention for same column asteroids:
			// slope : min value of float64
			slope := math.SmallestNonzeroFloat64
			if sameSlope, ok := findListForSlope(slope, sameSlopeList); ok {
				sameSlope.points = append(sameSlope.points, p2)
			} else {
				sameSlopeList = append(sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
			}
		} else {
			slope := getSlope(p1, p2)
			if sameSlope, ok := findListForSlope(slope, sameSlopeList); ok {
				sameSlope.points = append(sameSlope.points, p2)
			} else {
				sameSlopeList = append(sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
			}
		}
	}
	// remove asteroids that are blocked and cannot be seen
	removeAsteroidsThatAreBlockedForPoint(p1, sameSlopeList)
	// sort the asteroids based on slope descending value
	sort.Slice(sameSlopeList, func(i, j int) bool {
		return cmp(sameSlopeList[i].slope, sameSlopeList[j].slope) == 1
	})

	for _, s := range sameSlopeList {
		asteroidsInRange = append(asteroidsInRange, s.points...)
	}
	return asteroidsInRange
}

func buildSameSlopeList(p1, p2 Point, slope float64, sameSlopeList *[]SameSlope) []SameSlope {
	if sameSlope, ok := findListForSlope(slope, *sameSlopeList); ok {
		sameSlope.points = append(sameSlope.points, p2)
	} else {
		*sameSlopeList = append(*sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
	}
	return *sameSlopeList
}

func removeAsteroidsThatAreBlockedForPoint(from Point, someSlopeList []SameSlope) {
	for i, s := range someSlopeList {
		// for each reachable list of asteroids we try to exclude the blocked ones,
		// we do that by looking for three patterns:
		//     - middle: x 0 x
		//     - left  : 0 x x
		//     - right : x x 0
		// We iterate over all combinations for each case, if we don't find any match for
		// first pattern, we continue with the next pattern
		for position := Middle; position != Right+1; position++ {
			var newPoints []Point
			if len(s.points) > 1 {
				minDistance := math.MaxFloat64
				for _, p1 := range s.points {
					for _, p2 := range s.points {
						if p1 == p2 {
							continue
						}
						switch position {
						case Middle:
							if dist := distance(p1, p2); isBetween(p1, from, p2) && cmp(dist, minDistance) == -1 {
								newPoints = nil
								minDistance = dist
								newPoints = append(newPoints, p1)
								newPoints = append(newPoints, p2)
							}
						case Left:
							if dist := distance(p1, p2); isBetween(from, p1, p2) && cmp(dist, minDistance) == -1 {
								newPoints = nil
								minDistance = dist
								newPoints = append(newPoints, p1)
							}
						case Right:
							if dist := distance(p1, p2); isBetween(p1, p2, from) && cmp(dist, minDistance) == -1 {
								newPoints = nil
								minDistance = dist
								newPoints = append(newPoints, p2)
							}
						}
					}
				}
			}
			if len(newPoints) > 0 {
				someSlopeList[i].points = newPoints
				break
			}
		}
	}
}

func getAsteroidsForQuadrant(quadrant Quadrant, monitoringStation Point, asteroids []Point) []Point {
	var quadrantPoints []Point
	p := monitoringStation
	for _, a := range asteroids {
		if p == a {
			continue
		}
		switch quadrant {
		case One:
			if a.x < p.x && a.y >= p.y {
				quadrantPoints = append(quadrantPoints, a)
			}
		case Four:
			if a.x >= p.x && a.y > p.y {
				quadrantPoints = append(quadrantPoints, a)
			}
		case Three:
			if a.x > p.x && a.y <= p.y {
				quadrantPoints = append(quadrantPoints, a)
			}
		case Two:
			if a.x <= p.x && a.y < p.y {
				quadrantPoints = append(quadrantPoints, a)
			}
		}
	}

	return quadrantPoints
}

func part1(grid [][]int) (int, Point) {
	var maxPoint Point
	asteroids := getListOfAsteroids(grid)
	maxValue := math.MinInt32
	// iterate over each asteroid and check how many other asteroids it can detect
	for _, p := range asteroids {
		asteroidsInRange := getAsteroidsInRageForStation(p, asteroids)
		if len(asteroidsInRange) > maxValue {
			maxValue = len(asteroidsInRange)
			maxPoint = p
		}
	}

	return maxValue, maxPoint
}

func part2(grid [][]int, monitoringStation Point) int {
	count := 1
	result := 0
	asteroids := getListOfAsteroids(grid)
	numberOfAsteroids := len(asteroids)
	var asteroidsReached []Point
outerloop:
	for numberOfAsteroids > 1 {
		for _, quadrant := range []Quadrant{One, Four, Three, Two} {
			points := getAsteroidsForQuadrant(quadrant, monitoringStation, asteroids)
			asteroidsReached = getAsteroidsInRageForStation(monitoringStation, points)
			for _, p := range asteroidsReached {
				grid[p.x][p.y] = 0
				if count == 200 {
					result = p.y*100 + p.x
					break outerloop
				}
				count++
				numberOfAsteroids--
			}
			asteroids = getListOfAsteroids(grid)
		}
	}

	return result
}

func main() {
	file, err := os.Open("input/part1.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	grid := make([][]int, len(lines))
	for i, line := range lines {
		grid[i] = make([]int, len(line))
		for j, c := range line {
			s := 0
			if c == '#' {
				s = 1
			}
			grid[i][j] = s
		}
	}

	// part 1 solution
	max, point := part1(grid)
	fmt.Println("The result to 1st part is: ", max)

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(grid, point))
}
