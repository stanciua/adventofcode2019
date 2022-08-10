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
	y int64
	x int64
}

type Neighbor struct {
	pos    Position
	symbol rune
	dir    int64
}

const (
	Wall   = 0
	Move   = 1
	Oxygen = 2
)

const (
	WallSymbol    = '#'
	KnownPosition = '.'
	OxygenSymbol  = 'O'
)

const (
	North int64 = 1
	South int64 = 2
	West  int64 = 3
	East  int64 = 4
)

var REVERSE_COMMAND = []int64{2, 1, 4, 3}

var DIRECTIONS = []Position{{y: -1, x: 0}, {y: 1, x: 0}, {y: 0, x: -1}, {y: 0, x: 1}}

type Droid struct {
	vm   *VM
	area map[Position]rune
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
	area := make(map[Position]rune)
	droid := Droid{vm: NewVM(), area: area}
	droid.vm.loadProgram(input)
	startPosition := Position{y: 0, x: 0}
	discovered := make(map[Position]bool)
	oxygenPos := Position{y: 0, x: 0}
	// build the map and find the Oxygen position
	droid.buildMap(startPosition, discovered, &oxygenPos)
	droid.area[startPosition] = KnownPosition
	droid.area[oxygenPos] = KnownPosition
	return droid.findMininumNoOfSteps(startPosition, oxygenPos)
}

func part2(input []int64) int {
	area := make(map[Position]rune)
	droid := Droid{vm: NewVM(), area: area}
	droid.vm.loadProgram(input)
	startPosition := Position{y: 0, x: 0}
	discovered := make(map[Position]bool)
	oxygenPos := Position{y: 0, x: 0}
	// build the map and find the Oxygen position
	droid.buildMap(startPosition, discovered, &oxygenPos)
	droid.area[startPosition] = KnownPosition
	droid.area[oxygenPos] = KnownPosition
	return droid.fillWithOxygen(oxygenPos)
}

func getMinValue(queue map[Position]bool, dist map[Position]int) (Position, int) {
	min := math.MaxInt32
	pos := Position{y: 0, x: 0}
	for p, d := range dist {
		if d < min && queue[p] {
			min = d
			pos = p
		}
	}

	return pos, min
}

func (droid *Droid) findMininumNoOfSteps(source Position, destination Position) int {
	dist := make(map[Position]int)
	queue := make(map[Position]bool)

	for pos, sym := range droid.area {
		if sym != KnownPosition {
			continue
		}

		dist[pos] = math.MaxInt32
		queue[pos] = true
	}

	dist[source] = 0

	for len(queue) > 0 {
		// get u with min dist[u]
		minPos, minDist := getMinValue(queue, dist)

		if destination == minPos {
			return minDist
		}
		delete(queue, minPos)

		neighbors := droid.cellNeighbors(minPos)

		for _, n := range neighbors {
			if _, ok := queue[n]; !ok {
				continue
			}

			alt := minDist + 1
			if alt < dist[n] && minDist != math.MaxInt32 {
				dist[n] = alt
			}
		}
	}

	return 0
}

func (droid *Droid) fillWithOxygen(source Position) int {
	droid.area[source] = OxygenSymbol

	// count the number of KnownPositions
	count := 0
	for _, sym := range droid.area {
		if sym == KnownPosition {
			count++
		}
	}

	explored := make(map[Position]bool)
	minutes := 0
	for count > 0 {
		openLocations := make(map[Position]rune)
		for pos, sym := range droid.area {
			if sym == OxygenSymbol {
				for _, n := range droid.cellNeighbors(pos) {
					openLocations[n] = KnownPosition
				}
			}
		}

		// now fill every open location we found
		for pos := range openLocations {
			droid.area[pos] = OxygenSymbol
			explored[pos] = true
			count--
		}
		minutes++
	}

	return minutes
}

func (droid *Droid) cellNeighbors(pos Position) []Position {
	neighbors := make([]Position, 0)

	// check up
	p := Position{y: pos.y - 1, x: pos.x}
	if droid.area[p] == KnownPosition {
		neighbors = append(neighbors, p)
	}
	// check down
	p = Position{y: pos.y + 1, x: pos.x}
	if droid.area[p] == KnownPosition {
		neighbors = append(neighbors, p)
	}

	// check left
	p = Position{y: pos.y, x: pos.x - 1}
	if droid.area[p] == KnownPosition {
		neighbors = append(neighbors, p)
	}

	// check right
	p = Position{y: pos.y, x: pos.x + 1}
	if droid.area[p] == KnownPosition {
		neighbors = append(neighbors, p)
	}

	return neighbors

}
func (droid *Droid) droidStatusReply(move int64) int64 {
	output := int64(0)
	// execute instruction as long as we don't have any output, input or the program is done
	for {
		droid.vm.currInstruction = droid.vm.decodeCurrentInstruction()
		if droid.vm.currInstruction.opcode == Input {
			droid.vm.input = []int64{move}
			droid.vm.output = 0
		}

		droid.vm.executeCurrentInstruction()

		if droid.vm.hasFinished() || droid.vm.outputReady {
			output = droid.vm.output
			droid.vm.outputReady = false
			break
		}
	}

	return output
}

func (droid *Droid) buildMap(source Position, discovered map[Position]bool, oxygenPos *Position) {
	discovered[source] = true

	neighbors := droid.findNeighbors(source)

	for _, neighbor := range neighbors {
		if neighbor.symbol == OxygenSymbol {
			*oxygenPos = neighbor.pos
		} else if neighbor.symbol == WallSymbol {
			// mark all walls as already discovered
			discovered[neighbor.pos] = true
		}
		droid.area[neighbor.pos] = neighbor.symbol
	}

	// now visit all the cells that have not been visited
	for _, neighbor := range neighbors {
		if discovered[neighbor.pos] {
			continue
		}

		// we also need to move the droid in that direction
		_ = droid.droidStatusReply(neighbor.dir)

		// before moving the droid, mark the cell as known location
		droid.area[source] = KnownPosition
		droid.buildMap(neighbor.pos, discovered, oxygenPos)
		// if we need to go back, we need to tell the droid to backtrack
		_ = droid.droidStatusReply(REVERSE_COMMAND[neighbor.dir-1])
	}
}

func (droid *Droid) findNeighbors(pos Position) []Neighbor {
	neighbors := make([]Neighbor, 4)

	for d := North; d <= East; d++ {
		cPos := DIRECTIONS[d-1]
		dPos := Position{y: pos.y + cPos.y, x: pos.x + cPos.x}
		neighbors[d-1].pos = dPos
		neighbors[d-1].dir = d
		// pass the coordinate to the robot and check their response
		output := droid.droidStatusReply(d)

		if output == Wall {
			neighbors[d-1].symbol = WallSymbol
		} else if output == Move {
			neighbors[d-1].symbol = KnownPosition
			// go back to previous position
			_ = droid.droidStatusReply(REVERSE_COMMAND[d-1])
		} else {
			// found the oxygen system
			neighbors[d-1].symbol = OxygenSymbol
			// go back to previous position
			_ = droid.droidStatusReply(REVERSE_COMMAND[d-1])
		}
	}

	return neighbors
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
