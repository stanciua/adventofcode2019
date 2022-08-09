package main

import (
	"bufio"
	"fmt"
	"log"
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
	NoCmd  int64 = -1
	Wall         = 0
	Move         = 1
	Oxygen       = 2
)

const (
	DroidSymbol        rune = 'D'
	WallSymbol              = '#'
	KnownPosition           = '.'
	UnexploredPosition      = ' '
	OxygenSymbol            = 'O'
)

const (
	North int64 = 1
	South       = 2
	West        = 3
	East        = 4
)

const (
	MAX_Y = 50
	MAX_X = 50
)

var REVERSE_COMMAND = []int64{2, 1, 4, 3}

var COORDINATE_DIRECTIONS = []Position{{y: -1, x: 0}, {y: 1, x: 0}, {y: 0, x: -1}, {y: 0, x: 1}}
var REVERSE_COORDINATE_DIRECTIONS = []Position{{y: 1, x: 0}, {y: -1, x: 0}, {y: 0, x: 1}, {y: 0, x: -1}}

type Droid struct {
	vm      *VM
	area    map[Position]rune
	visited map[Position]bool
	path    map[Position]bool
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
	visited := make(map[Position]bool)
	path := make(map[Position]bool)
	droid := Droid{vm: NewVM(), area: area, visited: visited, path: path}
	// load the program into the cabinet memory
	droid.vm.loadProgram(input)
	startPosition := Position{y: 0, x: 0}
	discovered := make(map[Position]bool)
	// // initialize the area to spaces
	// for i := -25; i < 25; i++ {
	// 	for j := -25; j < 25; j++ {
	// 		droid.area[Position{y: int64(i), x: int64(j)}] = UnexploredPosition
	// 	}
	// }
	pos := droid.findOxygenSystem(startPosition, discovered)
	fmt.Println(pos)
	// fmt.Println(droid.area)
	// fmt.Println(discovered)
	droid.area[startPosition] = 'S'
	droid.plotMap()
	return 0
}

func part2(input []int64) int64 {
	return -1
}

func (droid *Droid) getOutput(move int64) (output int64, done bool) {
	// execute instruction as long as we don't have any output, input or the program is done
	for {
		droid.vm.currInstruction = droid.vm.decodeCurrentInstruction()
		if droid.vm.currInstruction.opcode == Input {
			droid.vm.input = []int64{move}
			droid.vm.output = 0
		}
		droid.vm.executeCurrentInstruction()
		if droid.vm.hasFinished() {
			done = true
			break
		}
		if droid.vm.outputReady {
			output = droid.vm.output
			droid.vm.outputReady = false
			done = false
			break
		}
	}

	return output, done
}

func getNextPositionAfterCommand(pos Position, command int64) Position {
	dirPos := COORDINATE_DIRECTIONS[command-1]
	return Position{y: pos.y + dirPos.y, x: pos.x + dirPos.x}
}

func (droid *Droid) backtrackToPositionIfDeadEnd(pos Position) (int64, Position, bool) {
	count := 0
	backtrackPos := Position{y: 0, x: 0}
	dir := North
	for d := North; d <= East; d++ {
		currPos := getNextPositionAfterCommand(pos, d)
		if droid.area[currPos] == WallSymbol {
			count++
		} else if droid.area[currPos] == KnownPosition {
			dir = d
			backtrackPos = currPos
		}
	}

	// dead end if count is 3
	if count == 3 {
		return dir, backtrackPos, true
	}
	return dir, backtrackPos, false
}

func (droid *Droid) positionToMovementCommand(pos Position) int {
	movementCommand := 0
	if pos == (Position{y: -1, x: 0}) {
		movementCommand = 1
	} else if pos == (Position{y: 1, x: 0}) {
		movementCommand = 2
	} else if pos == (Position{y: 0, x: -1}) {
		movementCommand = 3
	} else {
		movementCommand = 4
	}

	return movementCommand
}

func (droid *Droid) findOxygenSystem(source Position, discovered map[Position]bool) Position {
	// droid.area[source] = DroidSymbol
	discovered[source] = true

	neighbors := droid.findNeighbors(source)

	for _, neighbor := range neighbors {
		if neighbor.symbol == OxygenSymbol {
			fmt.Println("FOUND THE Oxygen")
			// droid.area[neighbor.pos] = OxygenSymbol
			fmt.Println(neighbor.pos)
			// return neighbor.pos
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
		output, _ := droid.getOutput(neighbor.dir)

		if output == 1 {
		}

		// before moving the droid, mark the cell as known location
		droid.area[source] = KnownPosition
		_ = droid.findOxygenSystem(neighbor.pos, discovered)
		// if we need to go back, we need to tell the droid to backtrack
		output, _ = droid.getOutput(REVERSE_COMMAND[neighbor.dir-1])
		if output == 1 {
		}
		// return pos
	}
	// droid.area[source] = OxygenSymbol
	// count := 0
	// countTotal := 0
	// for pos, symbol := range droid.area {
	// 	if symbol != KnownPosition {
	// 		continue
	// 	}
	// 	count++
	// 	countTotal++
	//
	// 	neighbors := droid.findNeighbors(pos)
	// 	for _, n := range neighbors {
	// 		if n.symbol == UnexploredPosition {
	// 			count--
	// 			break
	// 		}
	// 	}
	// }
	// fmt.Println("count: ", count)
	// fmt.Println("countTotal: ", count)
	return source
}

func (droid *Droid) plotMap() {
	var mapDisplay [MAX_Y][MAX_X]rune
	for i := 0; i < MAX_Y; i++ {
		for j := 0; j < MAX_X; j++ {
			mapDisplay[i][j] = ' '
		}
	}

	for pos, symbol := range droid.area {
		mapDisplay[MAX_Y/2+pos.y][MAX_X/2+pos.x] = symbol
	}

	mapDisplay[25+14][25-15] = OxygenSymbol
	for i := 0; i < MAX_Y; i++ {
		for j := 0; j < MAX_X; j++ {
			fmt.Print(string(mapDisplay[i][j]))
		}
		fmt.Println()
	}
}

func (droid *Droid) findNeighbors(pos Position) []Neighbor {
	neighbors := make([]Neighbor, 4)

	for d := North; d <= East; d++ {
		cPos := COORDINATE_DIRECTIONS[d-1]
		dPos := Position{y: pos.y + cPos.y, x: pos.x + cPos.x}
		neighbors[d-1].pos = dPos
		neighbors[d-1].dir = d
		// pass the coordinate to the robot and check their response
		output, _ := droid.getOutput(d)

		if output == Wall {
			neighbors[d-1].symbol = WallSymbol
		} else if output == Move {
			neighbors[d-1].symbol = KnownPosition
			// go back to previous position
			output, _ = droid.getOutput(REVERSE_COMMAND[d-1])
		} else {
			// found the oxygen system
			neighbors[d-1].symbol = OxygenSymbol
			// go back to previous position
			output, _ = droid.getOutput(REVERSE_COMMAND[d-1])
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
