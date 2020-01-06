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
	x int
	y int
}

type Symbol rune

const (
	Black Symbol = '.'
	White        = '#'
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

type Robot struct {
	brain         *VM
	region        map[Position]Symbol
	currPosition  Position
	currDirection Direction
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
	robot := Robot{brain: NewVM(), region: make(map[Position](Symbol))}
	// load the program into the robot memory
	robot.brain.loadProgram(input)
	robot.region[robot.currPosition] = Black
	robot.currDirection = DirUp
	robot.paint()
	return len(robot.region)
}

func part2(input []int64) {
	robot := Robot{brain: NewVM(), region: make(map[Position](Symbol))}
	// load the program into the robot memory
	robot.brain.loadProgram(input)
	robot.region[robot.currPosition] = White
	robot.currDirection = DirUp
	robot.paint()
	grid := make([][]Symbol, 256)
	for i := 0; i < 256; i++ {
		grid[i] = make([]Symbol, 256)
		for j := 0; j < 256; j++ {
			grid[i][j] = Black
		}
	}

	curr := Position{x: len(grid) / 2, y: len(grid) / 2}
	for k, v := range robot.region {
		grid[curr.x+k.x][curr.y+k.y] = v
	}

	displayGrid(grid)
}

func displayGrid(grid [][]Symbol) {
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			if grid[i][j] == Black {
				fmt.Print(".")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}

}

func (r *Robot) paint() {
	// we need to distinguish between output paint color and next direction turn
	paintColorFlg := true
	nextTurnFlg := false
	paintColor := Black
	nextTurn := DirLeft
	for !r.brain.hasFinished() {
		if v, ok := r.region[r.currPosition]; v == Black || !ok {
			r.brain.input = []int64{0}
		} else {
			r.brain.input = []int64{1}
		}
		r.brain.currInstruction = r.brain.decodeCurrentInstruction()
		r.brain.executeCurrentInstruction()
		if paintColorFlg && r.brain.outputReady {
			r.brain.outputReady = false
			if r.brain.output == 1 {
				paintColor = White
			} else {
				paintColor = Black
			}
			paintColorFlg = false
			nextTurnFlg = true
			continue
		} else if nextTurnFlg && r.brain.outputReady {
			r.brain.outputReady = false
			if r.brain.output == 1 {
				nextTurn = DirRight
			} else {
				nextTurn = DirLeft
			}
			paintColorFlg = true
			nextTurnFlg = false
		} else {
			continue
		}
		// paint the current position
		r.region[r.currPosition] = paintColor
		// get the new current position based on the output from the robot
		r.currPosition = r.getNextPosition(nextTurn)
		r.currDirection = r.getNextDirection(nextTurn)
	}
}

func (r *Robot) getNextPosition(nextTurn Direction) Position {
	newPos := r.currPosition
	switch r.currDirection {
	case DirUp:
		if nextTurn == DirLeft {
			newPos.y -= 1
		} else {
			newPos.y += 1
		}
	case DirDown:
		if nextTurn == DirLeft {
			newPos.y += 1
		} else {
			newPos.y -= 1
		}
	case DirLeft:
		if nextTurn == DirLeft {
			newPos.x += 1
		} else {
			newPos.x -= 1
		}
	case DirRight:
		if nextTurn == DirLeft {
			newPos.x -= 1
		} else {
			newPos.x += 1
		}
	default:
		panic("Invalid direction received!")
	}

	return newPos
}

func (r *Robot) getNextDirection(nextTurn Direction) Direction {
	newDirection := DirUp
	switch r.currDirection {
	case DirUp:
		if nextTurn == DirLeft {
			newDirection = DirLeft
		} else {
			newDirection = DirRight
		}
	case DirDown:
		if nextTurn == DirLeft {
			newDirection = DirRight
		} else {
			newDirection = DirLeft
		}
	case DirLeft:
		if nextTurn == DirLeft {
			newDirection = DirDown
		} else {
			newDirection = DirUp
		}
	case DirRight:
		if nextTurn == DirLeft {
			newDirection = DirUp
		} else {
			newDirection = DirDown
		}
	default:
		panic("Invalid direction received!")
	}

	return newDirection
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

	// fmt.Println("The result to 1st part is: ", part1(program))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ")
	part2(program)
}
