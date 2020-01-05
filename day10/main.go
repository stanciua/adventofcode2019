package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"sort"
)

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

	fmt.Println("The result to 1st part is: ", part1(grid))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(grid))
}

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

const epsilon = 1e-9

func cmp(a, b float64) int {
	if a-b < epsilon && math.Abs(a-b) > epsilon {
		return -1
	} else if a-b > epsilon && math.Abs(a-b) > epsilon {
		return 1
	}
	return 0
}

func getSlope(p1, p2 Point) *big.Float {
	return big.NewFloat(float64(p2.y-p1.y) / float64(p2.x-p1.x))
}

func distance(p1, p2 Point) *big.Float {
	return big.NewFloat(math.Sqrt(math.Pow(float64(p2.x-p1.x), 2) + math.Pow(float64(p2.y-p1.y), 2)))
}

func isBetween(p1, p2, p3 Point) bool {
	// return distance(p1, p2).SetPrec(20).SetMode(big.ToNearestEven).Add(distance(p2, p3).SetPrec(20).SetMode(big.ToNearestEven)).SetPrec(20).SetMode(big.ToNearestEven).Cmp(distance(p1, p3).SetPrec(20).SetMode(big.ToNearestEven))
	distp1p2 := distance(p1, p2)
	distp1p2 = distp1p2.SetPrec(20).SetMode(big.ToNearestEven)
	// fmt.Print("distp1p2: ")
	// fmt.Println(distp1p2.Float64())
	distp2p3 := distance(p2, p3)
	distp2p3 = distp2p3.SetPrec(20).SetMode(big.ToNearestEven)
	// fmt.Print("distp2p3: ")
	// fmt.Println(distp2p3.Float64())
	distp1p3 := distance(p1, p3)
	distp1p3 = distp1p3.SetPrec(20).SetMode(big.ToNearestEven)
	// fmt.Print("distp1p3: ")
	// fmt.Println(distp1p3.Float64())
	add := big.NewFloat(0).Add(distp1p2, distp2p3)
	add = add.SetPrec(20).SetMode(big.ToNearestEven)
	// fmt.Print("add: ")
	// fmt.Println(add.Float64())
	// fmt.Println(add.Cmp(distp1p3))
	return add.Cmp(distp1p3) == 0
}

func closestPoint(p Point, points []Point) Point {
	var point Point
	dist := big.NewFloat(math.MaxFloat64)
	for _, point := range points {
		d := distance(p, point)
		if d.Cmp(dist) == -1 {
			dist = d
			point = p
		}
	}
	return point
}

type SameSlope struct {
	points []Point
	slope  *big.Float
}

func findListForSlope(slope *big.Float, sameSlopeList []SameSlope) (*SameSlope, bool) {
	var sameSlope *SameSlope
	for i, v := range sameSlopeList {
		if v.slope.Cmp(slope) == 0 {
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
			slope := big.NewFloat(math.MaxFloat64)
			if sameSlope, ok := findListForSlope(slope, sameSlopeList); ok {
				sameSlope.points = append(sameSlope.points, p2)
			} else {
				sameSlopeList = append(sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
			}
		} else if p2.y-p1.y == 0 {
			// convention for same column asteroids:
			// slope : min value of float64
			slope := big.NewFloat(math.SmallestNonzeroFloat64)
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
	removeAsteroidsThatAreBlocked(p1, sameSlopeList)
	// sort the asteroids based on slope descending value
	sort.Slice(sameSlopeList, func(i, j int) bool {
		return sameSlopeList[i].slope.Cmp(sameSlopeList[j].slope) == 1
	})
	for _, s := range sameSlopeList {
		asteroidsInRange = append(asteroidsInRange, s.points...)
	}
	return asteroidsInRange
}

func part1(grid [][]int) int {
	asteroids := getListOfAsteroids(grid)
	var sameSlopeList []SameSlope
	// iterate over each asteroid and check how many other asteroids it can detect
	permutations := getFromToLocations(len(asteroids))
	maxValue := math.MinInt32
	for i := 0; i < len(permutations); i++ {
		p1 := asteroids[permutations[i][0]]
		p2 := asteroids[permutations[i][1]]
		if p2.x-p1.x == 0 {
			// convention for same line asteroids:
			// slope : max value of float64
			slope := big.NewFloat(math.MaxFloat64)
			if sameSlope, ok := findListForSlope(slope, sameSlopeList); ok {
				sameSlope.points = append(sameSlope.points, p2)
			} else {
				sameSlopeList = append(sameSlopeList, SameSlope{points: []Point{p2}, slope: slope})
			}

		} else if p2.y-p1.y == 0 {
			// convention for same column asteroids:
			// slope : min value of float64
			slope := big.NewFloat(math.SmallestNonzeroFloat64)
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

		// filter out asteroids that have the visibility blocked
		if i != 0 && (i+1)%(len(asteroids)-1) == 0 {
			removeAsteroidsThatAreBlocked(p1, sameSlopeList)
			sum := 0
			for _, sameSlope := range sameSlopeList {
				sum += len(sameSlope.points)

			}
			if sum > maxValue {
				maxValue = sum
			}
			sameSlopeList = nil
		}
	}
	return maxValue
}

func isPointInList(p Point, list []Point) bool {
	for _, v := range list {
		if p == v {
			return true
		}
	}

	return false
}

func removeAsteroidsThatAreBlocked(from Point, someSlopeList []SameSlope) {
	for i, s := range someSlopeList {
		// for each reachable list of asteroids we try to exclude the blocked one,
		// we do that looking for free cases:
		//     - middle: x 0 x
		//     - left  : 0 x x
		//     - right : x x 0
		// We iterate all combinations for each case, if we don't find any match for
		// first case, we continue with the next pattern
		permutations := getFromToLocations(len(s.points))
		for position := Middle; position != Right+1; position++ {
			newPoints := checkAsteroidsPositions(position, permutations, from, s.points)
			if len(newPoints) > 0 {
				plist := &someSlopeList[i].points
				*plist = newPoints
				break
			}
		}
	}
}

type Position int

const (
	Middle Position = iota
	Left
	Right
)

func checkAsteroidsPositions(position Position, permutations [][]int, from Point, points []Point) []Point {
	var newPoints []Point
	if len(points) > 1 {
		minDistance := big.NewFloat(math.MaxFloat64)
		for i := 0; i < len(permutations); i++ {
			p1 := points[permutations[i][0]]
			p2 := points[permutations[i][1]]
			var dist *big.Float
			cond := false
			switch position {
			case Middle:
				dist = distance(p1, p2)
				cond = isBetween(p1, from, p2)
				if cond && dist.Cmp(minDistance) == -1 {
					newPoints = nil
					minDistance = dist
					newPoints = append(newPoints, p1)
					newPoints = append(newPoints, p2)
				}
			case Left:
				dist = distance(from, p1)
				cond = isBetween(from, p1, p2)
				if cond && dist.Cmp(minDistance) == -1 {
					newPoints = nil
					minDistance = dist
					newPoints = append(newPoints, p1)
				}
			case Right:
				dist = distance(p2, from)
				cond = isBetween(p1, p2, from)
				if cond && dist.Cmp(minDistance) == -1 {
					newPoints = nil
					minDistance = dist
					newPoints = append(newPoints, p2)
				}
			}
		}

	}
	return newPoints
}

func getFromToLocations(size int) [][]int {
	list := make([]int, size)
	for i := 0; i < size; i++ {
		list[i] = i
	}
	var locations [][]int
	for i := 0; i < size; i++ {
		var newList []int
		newList = append(newList, list[:i]...)
		newList = append(newList, list[i+1:]...)
		for j := 0; j < len(newList); j++ {
			locations = append(locations, []int{list[i], newList[j]})
		}
	}
	return locations
}

func part2(grid [][]int) int {
	count := 1
	monitoringStation := Point{x: 29, y: 28}
	asteroids := getListOfAsteroids(grid)
	numberOfAsteroids := len(asteroids)
	var asteroidsReached []Point
	for numberOfAsteroids > 1 {
		for _, quadrant := range []Quadrant{One, Four, Three, Two} {
			points := getAsteroidsForQuadrant(quadrant, monitoringStation, asteroids)
			asteroidsReached = getAsteroidsInRageForStation(monitoringStation, points)
			for _, p := range asteroidsReached {
				grid[p.x][p.y] = 0
				if count == 200 {
					return p.y*100 + p.x
				}
				count++
				numberOfAsteroids--
			}
			asteroids = getListOfAsteroids(grid)
		}
	}
	return -1
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
