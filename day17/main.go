package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Point struct {
	y int
	x int
}

type Neighbor struct {
	pos    Point
	symbol rune
}

const (
	Scaffold     = '#'
	OpenSpace    = '.'
	NewLine      = '\n'
	RobotUp      = '^'
	RobotDown    = 'v'
	RobotLeft    = '<'
	RobotRight   = '>'
	Intersection = 'O'
)

type Robot struct {
	vm   *VM
	view [][]rune
}

type VM struct {
	memory             []int64
	currInstruction    *Instruction
	input              []int64
	instructionPointer int64
	output             int64
	relativeBase       int64
	outputReady        bool
}

type Instruction struct {
	opcode    Opcode
	paramMode []ParameterMode
	params    []int64
	length    int64
}

func NewVM() *VM {

	return new(VM)
}

func (vm *VM) loadProgram(input []int64) {
	vm.memory = append(vm.memory, input...)
}

type Opcode int64

const (
	Add Opcode = iota + 1
	Multiply
	Input
	Output
	JumpIfTrue
	JumpIfFalse
	LessThan
	Equals
	UpdateRelativeBase
)

type ParameterMode int64

const (
	Positional ParameterMode = iota
	Immediate
	Relative
)

const (
	Up int = iota
	Down
	Left
	Right
)

// UP, DOWN, LEFT, RIGHT
var DIRECTIONS = []Point{Point{-1, 0}, Point{1, 0}, Point{0, -1}, Point{0, 1}}

func (vm *VM) hasFinished() bool {
	return vm.load(vm.instructionPointer) == 99
}

func (vm *VM) decodeCurrentInstruction() *Instruction {
	instruction := new(Instruction)
	opcode := vm.load(vm.instructionPointer)
	instruction.opcode = Opcode(opcode % 100)
	switch Opcode(instruction.opcode) {
	case Add, Multiply, LessThan, Equals:
		instruction.params = []int64{vm.load(vm.instructionPointer + 1),
			vm.load(vm.instructionPointer + 2),
			vm.load(vm.instructionPointer + 3)}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10),
			ParameterMode(opcode / 1000 % 10),
			ParameterMode(opcode / 10000 % 10)}
		instruction.length = 4
	case Input, Output, UpdateRelativeBase:
		instruction.params = []int64{vm.load(vm.instructionPointer + 1)}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10)}
		instruction.length = 2
	case JumpIfTrue, JumpIfFalse:
		instruction.params = []int64{vm.load(vm.instructionPointer + 1),
			vm.load(vm.instructionPointer + 2)}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10),
			ParameterMode(opcode / 1000 % 10)}
		instruction.length = 3
	default:
		panic(fmt.Sprintf("Invalid instruction received: %d", instruction.opcode))
	}

	return instruction
}

func (vm *VM) load(address int64) int64 {
	for address > int64(len(vm.memory)-1) {
		vm.doubleMemory()
	}

	return vm.memory[address]
}

func (vm *VM) store(address int64, mode ParameterMode, val int64) {
	for address > int64(len(vm.memory)-1) || address+vm.relativeBase > int64(len(vm.memory)-1) {
		vm.doubleMemory()
	}

	if mode == Relative {
		vm.memory[vm.relativeBase+address] = val
	} else {
		vm.memory[address] = val
	}
}

func (vm *VM) doubleMemory() {
	vm.memory = append(vm.memory, make([]int64, len(vm.memory)*2)...)
}

func (vm *VM) getParamValue(index int64) int64 {
	value := int64(0)
	switch vm.currInstruction.paramMode[index] {
	case Positional:
		value = vm.load(vm.currInstruction.params[index])
	case Immediate:
		value = vm.currInstruction.params[index]
	case Relative:
		value = vm.load(vm.relativeBase + vm.currInstruction.params[index])
	}

	return value
}

func (vm *VM) executeCurrentInstruction() {
	switch vm.currInstruction.opcode {
	case Add:
		// add the input and store it inside the output
		vm.store(vm.currInstruction.params[2], vm.currInstruction.paramMode[2], vm.getParamValue(0)+vm.getParamValue(1))
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Multiply:
		// multiply the input and store it inside the output
		vm.store(vm.currInstruction.params[2], vm.currInstruction.paramMode[2], vm.getParamValue(0)*vm.getParamValue(1))
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Input:
		// store the input of the program
		vm.store(vm.currInstruction.params[0], vm.currInstruction.paramMode[0], vm.input[0])
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Output:
		// store the instruction output into the VM output and print it to the screen
		vm.output = vm.getParamValue(0)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
		vm.outputReady = true
	case JumpIfTrue:
		if vm.getParamValue(0) != 0 {
			vm.instructionPointer = vm.getParamValue(1)
		} else {
			// increment the instruction pointer and memory
			vm.instructionPointer += vm.currInstruction.length
		}
	case JumpIfFalse:
		if vm.getParamValue(0) == 0 {
			vm.instructionPointer = vm.getParamValue(1)
		} else {
			// increment the instruction pointer and memory
			vm.instructionPointer += vm.currInstruction.length
		}
	case LessThan:
		val := int64(0)
		if vm.getParamValue(0) < vm.getParamValue(1) {
			val = 1
		} else {
			val = 0
		}
		vm.store(vm.currInstruction.params[2], vm.currInstruction.paramMode[2], val)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Equals:
		val := int64(0)
		if vm.getParamValue(0) == vm.getParamValue(1) {
			val = 1
		} else {
			val = 0
		}
		vm.store(vm.currInstruction.params[2], vm.currInstruction.paramMode[2], val)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case UpdateRelativeBase:
		// update the relative base address
		vm.relativeBase += vm.getParamValue(0)
		vm.instructionPointer += vm.currInstruction.length
	default:
		panic("Unknown opcode encountered!")
	}
}

func part1(input []int64) int {
	view := make([][]rune, 0)
	robot := Robot{vm: NewVM(), view: view}
	robot.vm.loadProgram(input)
	// build the map and find the Oxygen position
	robot.buildView()
	intersectionPoints := robot.findIntersectionPoints()
	// robot.displayView()
	sum := 0
	for _, aligment := range robot.computeAligments(intersectionPoints) {
		sum += aligment
	}
	return sum
}

func part2(input []int64) int {
	view := make([][]rune, 0)
	robot := Robot{vm: NewVM(), view: view}
	robot.vm.loadProgram(input)
	// build the map and find the Oxygen position
	robot.buildView()
	robot.displayView()
	// start, end := robot.findStartEndPositions()
	// visited := make([]Point, 0)
	// currentPath := make([]Point, 0)
	// currentPath = append(currentPath, start)
	// path := make([]Point, 0)
	// visited := make(map[Point]bool)
	// prev := Point{0, 0}
	// scaffolds := robot.getScaffolds()
	// resultPath := make([]Point, 0)
	// robot.dfs(start, end, path, visited, prev, scaffolds, &resultPath)
	// fmt.Println(resultPath)
	// translatedPath := translatePath(resultPath, RobotUp)
	// fmt.Println(translatedPath)
	path := []string{"R", "8", "R", "8", "R", "4", "R", "4", "R", "8", "L", "6", "L", "2", "R", "4", "R", "4", "R", "8", "R", "8", "R", "8", "L", "6", "L", "2"}
	_, _, _ = getMovementFunctions(path)
	return 0
}

func findInPath(path []Point, p Point) bool {
	for _, pi := range path {
		if p == pi {
			return true
		}
	}

	return false
}
func (robot *Robot) robotCameraOutput() (done bool, output rune) {
	// execute instruction as long as we don't have any output, input or the program is done
	for {
		robot.vm.currInstruction = robot.vm.decodeCurrentInstruction()
		robot.vm.executeCurrentInstruction()
		if robot.vm.outputReady {
			output = rune(robot.vm.output)
			robot.vm.outputReady = false
			done = false
			break
		}
		if robot.vm.hasFinished() {
			done = true
			break
		}
	}

	return done, output
}

func (robot *Robot) findStartEndPositions() (start Point, end Point) {
	for i := 0; i < len(robot.view); i++ {
		for j := 0; j < len(robot.view[i]); j++ {
			if robot.view[i][j] == RobotUp {
				start = Point{i, j}
			} else if robot.view[i][j] == Scaffold && len(robot.findNeighbors(Point{i, j})) == 1 {
				end = Point{i, j}
			}
		}
	}

	return start, end
}

func (robot *Robot) isScaffold(p Point) bool {
	sym := robot.view[p.y][p.x]
	if sym == Scaffold || sym == RobotUp || sym == RobotDown || sym == RobotLeft || sym == RobotRight || sym == Intersection {
		return true
	}

	return false
}

func (robot *Robot) findNeighbors(p Point) []Point {
	neighbors := make([]Point, 0)
	if p.y-1 >= 0 && robot.isScaffold(Point{p.y - 1, p.x}) {
		neighbors = append(neighbors, Point{p.y - 1, p.x})
	}
	if p.y+1 < len(robot.view) && robot.isScaffold(Point{p.y + 1, p.x}) {
		neighbors = append(neighbors, Point{p.y + 1, p.x})
	}
	if p.x-1 >= 0 && robot.isScaffold(Point{p.y, p.x - 1}) {
		neighbors = append(neighbors, Point{p.y, p.x - 1})
	}
	if p.x+1 < len(robot.view[p.y]) && robot.isScaffold(Point{p.y, p.x + 1}) {
		neighbors = append(neighbors, Point{p.y, p.x + 1})
	}

	return neighbors
}

func (robot *Robot) dfs(start, end Point, path []Point, visited map[Point]bool, prev Point, scaffolds []Point, resultPath *[]Point) {
	// if we already found the one path, we just unwind the stack
	// if len(*resultPath) > 0 {
	// 	return
	// }

	path = append(path, start)
	visited[start] = true

	if start == end {
		// we are interested in only one path that passes
		// through all the scaffolds and ignore the rest
		if len(path) >= len(scaffolds) {
			validPath := true
			for _, s := range scaffolds {
				if !findInPath(path, s) {
					validPath = false
					break
				}
			}

			if validPath {
				*resultPath = path[:]
				fmt.Println(translatePath(*resultPath, RobotUp))
			}

		}
	} else {
		for _, p := range robot.findNeighbors(start) {
			if prev == p {
				continue
			}
			visitInter := false
			if visited[p] && robot.isIntersectionPoint(p) {
				// we need to look if there's an unexplored path from
				// intersection point
				for _, n := range robot.findNeighbors(p) {
					if !visited[n] {
						visitInter = true
						break
					}
				}
			}
			if !visited[p] || visitInter {
				prev = start
				robot.dfs(p, end, path, visited, prev, scaffolds, resultPath)
			}
		}
	}

	path = path[:len(path)-1]
	delete(visited, start)
}

func (robot *Robot) getScaffolds() []Point {
	scaffolds := make([]Point, 0)
	for i, line := range robot.view {
		for j, s := range line {
			if s == Scaffold || s == RobotUp {
				scaffolds = append(scaffolds, Point{i, j})
			}
		}
	}

	return scaffolds
}

func getDirectionFromPoint(p Point) int {
	for idx, d := range DIRECTIONS {
		if p == d {
			return idx
		}
	}
	panic("invalid value for point")
}

func translatePath(path []Point, robotDirection rune) []string {
	translatedPath := make([]string, 0)

	prev := path[0]
	path = path[1:]
	count := 0
	for _, p := range path {
		newDir := Point{p.y - prev.y, p.x - prev.x}
		switch robotDirection {
		case RobotUp:
			switch getDirectionFromPoint(newDir) {
			case Up:
				count++
			case Left:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "L")
				robotDirection = RobotLeft
			case Right:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "R")
				robotDirection = RobotRight
			}
		case RobotDown:
			switch getDirectionFromPoint(newDir) {
			case Down:
				count++
			case Left:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "R")
				robotDirection = RobotLeft
			case Right:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "L")
				robotDirection = RobotRight
			}
		case RobotLeft:
			switch getDirectionFromPoint(newDir) {
			case Left:
				count++
			case Up:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "R")
				robotDirection = RobotUp
			case Down:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "L")
				robotDirection = RobotDown
			}
		case RobotRight:
			switch getDirectionFromPoint(newDir) {
			case Right:
				count++
			case Up:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "L")
				robotDirection = RobotUp
			case Down:
				if count > 0 {
					translatedPath = append(translatedPath, strconv.Itoa(count+1))
					count = 0
				}
				translatedPath = append(translatedPath, "R")
				robotDirection = RobotDown
			}
		}
		prev = p
	}
	// if we end up here, we also need to add the last counter
	translatedPath = append(translatedPath, strconv.Itoa(count+1))

	return translatedPath
}

func getMovementFunctions(path []string) (a []string, b []string, c []string) {
	// A should be at the begining while C should be at the end
	// B should be somewhere in the middle
	a = make([]string, 0)
	b = make([]string, 0)
	c = make([]string, 0)

	current := make([]string, 0)
	for i := 0; i < len(path); i++ {
		current = append(current, path[i])
		// search maximum current length into the copyPath
		// and see if we can find a match
		if findSubPathOccurence(current, path) {
			continue
		} else {
			// we didnt find an occurence so stop here and find out what A is
			current = current[:len(current)-1]
			break
		}
	}

	fmt.Println(current)

	return a, b, c
}

func findSubPathOccurence(subPath []string, path []string) bool {
	for i, m := range subPath {
		if m == path[i] {
			continue
		} else {
			return false
		}
	}

	return true
}

func (robot *Robot) buildView() {
	// i := 0
	// robot.view = append(robot.view, make([]rune, 0))
	// for done, output := robot.robotCameraOutput(); !done; done, output = robot.robotCameraOutput() {
	// 	if output == NewLine {
	// 		i++
	// 		robot.view = append(robot.view, make([]rune, 0))
	// 	} else {
	// 		robot.view[i] = append(robot.view[i], output)
	// 	}
	// }
	//
	// // the robot sends two new lines at the end, remove them from the view
	// robot.view = robot.view[:len(robot.view)-2]

	// view := "###..\n#.#..\n..#..\n..#..\n^####\n..#.#\n..###"
	// view := "..#..........\n..#..........\n#######...###\n#.#...#...#.#\n#############\n..#...#...#..\n..#####...^.."
	view := "#######...#####\n#.....#...#...#\n#.....#...#...#\n......#...#...#\n......#...###.#\n......#.....#.#\n^########...#.#\n......#.#...#.#\n......#########\n........#...#..\n....#########..\n....#...#......\n....#...#......\n....#...#......\n....#####......"
	for idx, line := range strings.Split(view, "\n") {
		robot.view = append(robot.view, make([]rune, 0))
		for _, c := range line {
			robot.view[idx] = append(robot.view[idx], c)
		}
	}
}

func (robot *Robot) displayView() {
	for _, row := range robot.view {
		for _, col := range row {
			fmt.Print(string(col))
		}
		fmt.Println()
	}
}

func (robot *Robot) isIntersectionPoint(point Point) bool {
	up := point.y-1 >= 0 && robot.view[point.y-1][point.x] == Scaffold
	down := point.y+1 < len(robot.view) && robot.view[point.y+1][point.x] == Scaffold
	left := point.x-1 >= 0 && robot.view[point.y][point.x-1] == Scaffold
	right := point.x+1 < len(robot.view[point.y]) && robot.view[point.y][point.x+1] == Scaffold

	return up && down && left && right && robot.view[point.y][point.x] != OpenSpace

}

func (robot *Robot) findIntersectionPoints() map[Point]bool {
	intersections := make(map[Point]bool)
	for i := 0; i < len(robot.view); i++ {
		for j := 0; j < len(robot.view[i]); j++ {
			point := Point{y: i, x: j}
			if robot.isIntersectionPoint(point) {
				intersections[point] = true
				robot.view[i][j] = Intersection
			}
		}
	}

	return intersections
}

func (robot *Robot) computeAligments(intersectionPoints map[Point]bool) []int {
	aligments := make([]int, 0)

	for p := range intersectionPoints {
		aligments = append(aligments, p.y*p.x)
	}

	return aligments
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

	// we have just one line, with elements separated by commas
	if len(inputs) != 1 {
		panic("The input should be only one line long!")
	}

	var program []int64
	for _, integer := range strings.Split(inputs[0], ",") {
		if val, err := strconv.ParseInt(integer, 10, 64); err != nil {
			panic(err)
		} else {
			program = append(program, val)
		}
	}
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(program))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(program))
}
