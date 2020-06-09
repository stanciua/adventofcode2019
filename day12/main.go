package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type Position struct {
	x int64
	y int64
	z int64
}

type Speed struct {
	x int64
	y int64
	z int64
}

type Moon struct {
	position Position
	speed    Speed
}

func applyGravity(moons []Moon) {
	var combinations [][]Moon
	currentCombination := make([]Moon, 2)
	generateCombinations(moons, currentCombination, &combinations, 0, len(moons)-1, 0, 2)
	for i := range combinations {
		combinations[i][0].speed = Speed{x: 0, y: 0, z: 0}
		combinations[i][1].speed = Speed{x: 0, y: 0, z: 0}
		if combinations[i][0].position.x < combinations[i][1].position.x {
			combinations[i][0].speed.x++
			combinations[i][1].speed.x--
		} else if combinations[i][0].position.x > combinations[i][1].position.x {
			combinations[i][0].speed.x--
			combinations[i][1].speed.x++
		}
		if combinations[i][0].position.y < combinations[i][1].position.y {
			combinations[i][0].speed.y++
			combinations[i][1].speed.y--
		} else if combinations[i][0].position.y > combinations[i][1].position.y {
			combinations[i][0].speed.y--
			combinations[i][1].speed.y++
		}
		if combinations[i][0].position.z < combinations[i][1].position.z {
			combinations[i][0].speed.z++
			combinations[i][1].speed.z--
		} else if combinations[i][0].position.z > combinations[i][1].position.z {
			combinations[i][0].speed.z--
			combinations[i][1].speed.z++
		}
		// update speed of each moon in the pair
		updateMoonVelocity(combinations[i][0], moons)
		updateMoonVelocity(combinations[i][1], moons)
	}
}

func updateMoonVelocity(moon Moon, moons []Moon) {
	for i := range moons {
		if moon.position == moons[i].position {
			moons[i].speed.x += moon.speed.x
			moons[i].speed.y += moon.speed.y
			moons[i].speed.z += moon.speed.z
		}
	}
}

func updateMoonsPosition(moons []Moon) {
	for i := range moons {
		moons[i].position.x += moons[i].speed.x
		moons[i].position.y += moons[i].speed.y
		moons[i].position.z += moons[i].speed.z
	}
}
func generateCombinations(moons []Moon, currentCombination []Moon, combinations *[][]Moon, start int, end int, index int, width int) {
	if index == width {
		*combinations = append(*combinations, append([]Moon(nil), currentCombination...))
		return
	}

	for i := start; i <= end && end-i+1 >= width-index; i++ {
		currentCombination[index] = moons[i]
		generateCombinations(moons, currentCombination, combinations, i+1, end, index+1, width)
	}
}

func part1(moons []Moon) int {
	for i := 0; i < 1000; i++ {
		applyGravity(moons)
		updateMoonsPosition(moons)
	}

	// calculate the total energy for the moons
	sum := float64(0)
	for _, moon := range moons {
		sum += (math.Abs(float64(moon.position.x)) + math.Abs(float64(moon.position.y)) + math.Abs(float64(moon.position.z))) *
			(math.Abs(float64(moon.speed.x)) + math.Abs(float64(moon.speed.y)) + math.Abs(float64(moon.speed.z)))
	}
	return int(sum)
}

type PositionSpeed struct {
	p0 int64
	s0 int64
	p1 int64
	s1 int64
	p2 int64
	s2 int64
	p3 int64
	s3 int64
}

func part2(moons []Moon) int {
	steps := 0
	positionSpeedX := PositionSpeed{
		p0: moons[0].position.x,
		s0: moons[0].speed.x,
		p1: moons[1].position.x,
		s1: moons[1].speed.x,
		p2: moons[2].position.x,
		s2: moons[2].speed.x,
		p3: moons[3].position.x,
		s3: moons[3].speed.x,
	}

	positionSpeedY := PositionSpeed{
		p0: moons[0].position.y,
		s0: moons[0].speed.y,
		p1: moons[1].position.y,
		s1: moons[1].speed.y,
		p2: moons[2].position.y,
		s2: moons[2].speed.y,
		p3: moons[3].position.y,
		s3: moons[3].speed.y,
	}
	positionSpeedZ := PositionSpeed{
		p0: moons[0].position.z,
		s0: moons[0].speed.z,
		p1: moons[1].position.z,
		s1: moons[1].speed.z,
		p2: moons[2].position.z,
		s2: moons[2].speed.z,
		p3: moons[3].position.z,
		s3: moons[3].speed.z,
	}

	stepsX, stepsY, stepsZ := 0, 0, 0

	for {
		applyGravity(moons)
		updateMoonsPosition(moons)
		steps++
		psX := PositionSpeed{
			p0: moons[0].position.x,
			s0: moons[0].speed.x,
			p1: moons[1].position.x,
			s1: moons[1].speed.x,
			p2: moons[2].position.x,
			s2: moons[2].speed.x,
			p3: moons[3].position.x,
			s3: moons[3].speed.x,
		}

		psY := PositionSpeed{
			p0: moons[0].position.y,
			s0: moons[0].speed.y,
			p1: moons[1].position.y,
			s1: moons[1].speed.y,
			p2: moons[2].position.y,
			s2: moons[2].speed.y,
			p3: moons[3].position.y,
			s3: moons[3].speed.y,
		}
		psZ := PositionSpeed{
			p0: moons[0].position.z,
			s0: moons[0].speed.z,
			p1: moons[1].position.z,
			s1: moons[1].speed.z,
			p2: moons[2].position.z,
			s2: moons[2].speed.z,
			p3: moons[3].position.z,
			s3: moons[3].speed.z,
		}

		if stepsX == 0 && psX == positionSpeedX {
			stepsX = steps
		}
		if stepsY == 0 && psY == positionSpeedY {
			stepsY = steps
		}
		if stepsZ == 0 && psZ == positionSpeedZ {
			stepsZ = steps
		}
		if stepsX != 0 && stepsY != 0 && stepsZ != 0 {
			break
		}
	}
	return lcm(lcm(stepsX, stepsY), stepsZ)
}

func gcd(a int, b int) int {
	if a == 0 {
		return b
	}
	return gcd(b%a, a)
}

func lcm(a int, b int) int {
	return (a * b) / gcd(a, b)
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

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var moons []Moon
	for _, p := range inputs {
		coordinates := strings.Split(strings.Trim(p, "<>"), ",")
		p := [3]int64{0, 0, 0}
		for i, c := range coordinates {
			if val, err := strconv.ParseInt(strings.Trim(strings.Split(c, "=")[1], " "), 10, 64); err != nil {
				panic(err)
			} else {
				p[i] = val
			}
		}

		moons = append(moons, Moon{
			position: Position{
				x: p[0],
				y: p[1],
				z: p[2],
			},
			speed: Speed{
				x: 0,
				y: 0,
				z: 0,
			},
		})
	}

	moonsCopy := append([]Moon(nil), moons...)

	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(moonsCopy))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(moons))
}
