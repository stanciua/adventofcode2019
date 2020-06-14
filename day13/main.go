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
	x int64
	y int64
}

type Symbol rune

const (
	Empty            Symbol = '0'
	Wall                    = '1'
	Block                   = '2'
	HorizontalPaddle        = '3'
	Ball                    = '4'
)

type Cabinet struct {
	vm                    *VM
	screen                map[Position]Symbol
	currentPaddlePosition Position
	currentBallPosition   Position
	currentScore          int64
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
	cabinet := Cabinet{vm: NewVM(), screen: make(map[Position](Symbol))}
	// load the program into the cabinet memory
	cabinet.vm.loadProgram(input)
	cabinet.run()
	noBlocks := 0
	for _, tile := range cabinet.screen {
		if tile == Block {
			noBlocks++
		}
	}
	return noBlocks
}

func part2(input []int64) int64 {
	cabinet := Cabinet{vm: NewVM(), screen: make(map[Position](Symbol))}
	// load the program into the cabinet memory
	cabinet.vm.loadProgram(input)
	// insert coin
	cabinet.vm.memory[0] = 2
	cabinet.run()
	return cabinet.currentScore
}

func (c *Cabinet) getOutput() (output int64, done bool) {
	// execute instruction as long as we don't have any output, input or the cabinet is done
	for true {
		c.vm.currInstruction = c.vm.decodeCurrentInstruction()
		// special case when we need to provide the input instruction with how to move the paddle:
		//   -1: paddle left
		//    1: paddle right
		//    0: paddle neutral
		if c.vm.currInstruction.opcode == Input {
			if c.currentBallPosition.y < c.currentPaddlePosition.y {
				c.vm.input = []int64{-1}
			} else if c.currentBallPosition.y > c.currentPaddlePosition.y {
				c.vm.input = []int64{1}
			} else {
				c.vm.input = []int64{0}
			}
		}
		c.vm.executeCurrentInstruction()
		if c.vm.hasFinished() {
			done = true
			break
		}
		if c.vm.outputReady {
			output = c.vm.output
			c.vm.outputReady = false
			done = false
			break
		}
	}

	return output, done
}

func (c *Cabinet) run() {
	for true {
		y, done := c.getOutput()
		if done {
			return
		}
		x, done := c.getOutput()
		if done {
			return
		}
		id, done := c.getOutput()
		if done {
			return
		}
		// if id is paddle '3' we need to update the current paddle position
		if id == 3 {
			c.currentPaddlePosition = Position{x: x, y: y}
		}
		// if id is ball '4' we need to update the current ball position
		if id == 4 {
			c.currentBallPosition = Position{x: x, y: y}
		}
		// if x = -1 , y = 0, then id is in fact the score of the current game
		if y == -1 && x == 0 {
			c.currentScore = id
		} else {
			c.screen[Position{x: x, y: y}] = symbolFromID(id)
		}
	}
}

func symbolFromID(id int64) Symbol {
	var symbol Symbol
	switch id {
	case 0:
		symbol = Empty
	case 1:
		symbol = Wall
	case 2:
		symbol = Block
	case 3:
		symbol = HorizontalPaddle
	case 4:
		symbol = Ball
	default:
		panic("Invalid id value")
	}
	return symbol
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
