package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
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
	stopped            bool
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
	ProgramStop = 99
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
var DIRECTIONS = []Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

func (vm *VM) hasFinished() bool {
	return vm.stopped
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
	case ProgramStop:
		vm.stopped = true

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
	case ProgramStop:
		break

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
	sum := 0
	for _, aligment := range robot.computeAligments(intersectionPoints) {
		sum += aligment
	}
	return sum
}

func part2(input []int64) int {
	// instantiate a robot to find the paths
	view := make([][]rune, 0)
	robot := Robot{vm: NewVM(), view: view}
	robot.vm.loadProgram(input)
	robot.buildView()
	start, end := robot.findStartEndPositions()
	currentPath := make([]Point, 0)
	currentPath = append(currentPath, start)
	path := make([]Point, 0)
	visited := make(map[Point]bool)
	prev := Point{0, 0}
	scaffolds := robot.getScaffolds()
	paths := make([][]Point, 0)
	robot.findPaths(start, end, path, visited, prev, scaffolds, &paths)
	// wake up the robot
	robot = Robot{vm: NewVM(), view: view}
	robot.vm.loadProgram(input)
	robot.vm.memory[0] = 2
	output := int64(0)
	for _, p := range paths {
		translatedPath := translatePath(p, RobotUp)
		splitedPath := compressPathTo3Movements(translatedPath)
		if len(splitedPath) > 0 {
			main, a, b, c := getRoutines(splitedPath)
			output = robot.runVacuumRobot(main, a, b, c)
			break
		}
	}
	return int(output)
}

func (robot *Robot) runVacuumRobot(main []int64, a []int64, b []int64, c []int64) int64 {
	output := int64(0)
	idx := 0
	input := append([]int64(nil), main...)
	input = append(input, a...)
	input = append(input, b...)
	input = append(input, c...)
	input = append(input, []int64{int64('n'), int64('\n')}...)

	for {
		robot.vm.currInstruction = robot.vm.decodeCurrentInstruction()
		if robot.vm.currInstruction.opcode == Input {
			robot.vm.input = []int64{input[idx]}
			idx++
			robot.vm.output = 0
		}

		robot.vm.executeCurrentInstruction()

		if robot.vm.outputReady {
			robot.vm.outputReady = false
		}

		if robot.vm.hasFinished() {
			output = robot.vm.output
			break
		}
	}
	return output
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

func (robot *Robot) findPaths(start, end Point, path []Point, visited map[Point]bool, prev Point, scaffolds []Point, paths *[][]Point) {
	path = append(path, start)
	visited[start] = true

	if start == end {
		if len(path) >= len(scaffolds) {
			validPath := true
			for _, s := range scaffolds {
				if !findInPath(path, s) {
					validPath = false
					break
				}
			}

			if validPath {
				*paths = append(*paths, append([]Point(nil), path...))
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
				robot.findPaths(p, end, path, visited, prev, scaffolds, paths)
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

func compressPathTo3Movements(path []string) []string {
	occurences := make(map[string]bool)
	for i := 1; i <= len(path); i++ {
		copyPath := append([]string(nil), path...)
		for len(copyPath) > 0 && len(copyPath) >= i {
			key := strings.Join(copyPath[:i], "")
			occurences[key] = true
			copyPath = copyPath[1:]
		}
	}
	occurencesFiltered := make(map[string]bool)
	startCandidates := make(map[string]bool)
	for k := range occurences {
		// ignore keys that don't end in a number
		if _, err := strconv.Atoi(k[len(k)-1:]); err != nil {
			continue
		}
		// ignore keys that start with a number
		if _, err := strconv.Atoi(k[:1]); err == nil {
			continue
		}

		// ignore keys that are longer than 10 bytes as they
		// go past the 20 characters limit
		if isMovementTooLong(k) {
			continue
		}

		if strings.HasPrefix(strings.Join(path, ""), k) {
			startCandidates[k] = true
		}
		occurencesFiltered[k] = true
	}
	// put the start prefix candidates in a list and sort them
	// to get deterministic order
	prefixList := make([]string, 0)
	for k := range startCandidates {
		prefixList = append(prefixList, k)
	}
	// sort them
	sort.Sort(sort.StringSlice(prefixList))

	// do the same for all filtered paths in occurences map
	movementsList := make([]string, 0)
	for k := range occurencesFiltered {
		movementsList = append(movementsList, k)
	}
	// sort them
	sort.Sort(sort.StringSlice(movementsList))
	partition := make([]string, 0)
	for _, sp := range prefixList {
		currentPath := make([]string, 0)
		currentPath = append(currentPath, sp)
		compress(&currentPath, movementsList, path, &partition)
	}

	return partition
}

func getRoutines(path []string) (main []int64, a []int64, b []int64, c []int64) {
	movementFunctions := []rune{'A', 'B', 'C'}
	movements := make(map[string]rune)
	for _, m := range path {
		if _, ok := movements[m]; !ok {
			movements[m] = movementFunctions[0]
			movementFunctions = movementFunctions[1:]
		}
	}

	// main routine
	main = make([]int64, 0)
	for _, m := range path {
		main = append(main, int64(movements[m]))
		main = append(main, int64(','))
	}
	// append the newline
	main[len(main)-1] = int64('\n')

	// a, b, c routines
	a = make([]int64, 0)
	b = make([]int64, 0)
	c = make([]int64, 0)
	pd := false
	for m, mName := range movements {
		if mName == 'A' {
			for _, d := range m {
				if unicode.IsDigit(d) {
					if pd {
						a = a[:len(a)-1]
					}
					pd = true
				} else {
					pd = false
				}
				a = append(a, int64(d))
				a = append(a, int64(','))
			}
			a[len(a)-1] = int64('\n')
		} else if mName == 'B' {
			for _, d := range m {
				if unicode.IsDigit(d) {
					if pd {
						b = b[:len(b)-1]
					}
					pd = true
				} else {
					pd = false
				}
				b = append(b, int64(d))
				b = append(b, int64(','))
			}
			b[len(b)-1] = int64('\n')
		} else {
			for _, d := range m {
				if unicode.IsDigit(d) {
					if pd {
						c = c[:len(c)-1]
					}
					pd = true
				} else {
					pd = false
				}
				c = append(c, int64(d))
				c = append(c, int64(','))
			}
			c[len(c)-1] = int64('\n')
		}
	}
	return main, a, b, c
}

func isMovementTooLong(move string) bool {
	len := 0
	insideDigit := false
	for _, c := range move {
		if len > 10 {
			return true
		}
		if !unicode.IsDigit(c) {
			len++
			if insideDigit {
				len++
				insideDigit = false
			}
		} else {
			insideDigit = true
			continue
		}
	}
	return false
}

func compress(currentPath *[]string, movementsList []string, path []string, partition *[]string) {
	solution := make(map[string]bool)

	for _, s := range *currentPath {
		solution[s] = true
	}

	if len(solution) > 3 || len(*currentPath) > 10 {
		return
	}

	if strings.Join(*currentPath, "") == strings.Join(path, "") {
		// we should only have 3 movesets (A, B, C) and maximum of 10 movesets are
		// allowed in main routine
		if len(solution) == 3 && len(*currentPath) <= 10 {
			*partition = append(*partition, *currentPath...)
			return
		}
	}

	for _, m := range movementsList {
		*currentPath = append(*currentPath, m)

		if !strings.HasPrefix(strings.Join(path, ""), strings.Join(*currentPath, "")) {
			*currentPath = (*currentPath)[:len(*currentPath)-1]
			continue
		}
		compress(currentPath, movementsList, path, partition)
		*currentPath = (*currentPath)[:len(*currentPath)-1]
	}
}

func (robot *Robot) buildView() {
	i := 0
	robot.view = append(robot.view, make([]rune, 0))
	for done, output := robot.robotCameraOutput(); !done; done, output = robot.robotCameraOutput() {
		if output == NewLine {
			i++
			robot.view = append(robot.view, make([]rune, 0))
		} else {
			robot.view[i] = append(robot.view[i], output)
		}
	}

	// the robot sends two new lines at the end, remove them from the view
	robot.view = robot.view[:len(robot.view)-2]
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
