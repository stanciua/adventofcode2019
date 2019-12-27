package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
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

	// we have just one line, with elements separated by commas
	if len(inputs) != 1 {
		panic("The input should be only one line long!")
	}

	var program []int
	for _, integer := range strings.Split(inputs[0], ",") {
		if val, err := strconv.Atoi(integer); err != nil {
			panic(err)
		} else {
			program = append(program, val)
		}

	}

	programCopy := append(program[:0:0], program...)
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(programCopy))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(program))
}

func part1(input []int) int {
	vm := NewVM()
	vm.loadProgram(input)
	for vm.memory[vm.instructionPointer] != 99 {
		vm.currInstruction = vm.decodeCurrentInstruction()
		vm.executeCurrentInstruction()
	}

	return vm.output
}

func part2(input []int) int {
	vm := NewVM()
	vm.loadProgram(input)
	vm.input = 5
	for vm.memory[vm.instructionPointer] != 99 {
		vm.currInstruction = vm.decodeCurrentInstruction()
		vm.executeCurrentInstruction()
	}

	return vm.output
}

type VM struct {
	memory             []int
	instructionPointer int
	input              int
	output             int
	currInstruction    *Instruction
}

type Instruction struct {
	opcode    Opcode
	paramMode []ParameterMode
	params    []int
	length    int
}

func NewVM() *VM {
	return new(VM)
}

func (vm *VM) loadProgram(input []int) {
	vm.memory = append(vm.memory, input...)
	vm.input = 1
}

type Opcode int

const (
	Add Opcode = iota + 1
	Multiply
	Input
	Output
	JumpIfTrue
	JumpIfFalse
	LessThan
	Equals
)

type ParameterMode int

const (
	Positional ParameterMode = iota
	Immediate
)

func (vm *VM) decodeCurrentInstruction() *Instruction {
	instruction := new(Instruction)
	opcode := vm.memory[vm.instructionPointer]
	instruction.opcode = Opcode(opcode % 100)
	switch Opcode(instruction.opcode) {
	case Add, Multiply, LessThan, Equals:
		instruction.params = []int{vm.memory[vm.instructionPointer+1],
			vm.memory[vm.instructionPointer+2],
			vm.memory[vm.instructionPointer+3]}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10),
			ParameterMode(opcode / 1000 % 10),
			ParameterMode(opcode / 10000 % 10)}
		instruction.length = 4
	case Input, Output:
		instruction.params = []int{vm.memory[vm.instructionPointer+1]}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10)}
		instruction.length = 2
	case JumpIfTrue, JumpIfFalse:
		instruction.params = []int{vm.memory[vm.instructionPointer+1],
			vm.memory[vm.instructionPointer+2]}
		instruction.paramMode = []ParameterMode{ParameterMode(opcode / 100 % 10),
			ParameterMode(opcode / 1000 % 10)}
		instruction.length = 3
	default:
		panic(fmt.Sprintf("Invalid instruction received: %d", instruction.opcode))
	}

	return instruction
}

func (vm *VM) getParamValue(index int) int {
	value := 0
	switch vm.currInstruction.paramMode[index] {
	case Positional:
		value = vm.memory[vm.currInstruction.params[index]]
	case Immediate:
		value = vm.currInstruction.params[index]
	}

	return value
}

func (vm *VM) executeCurrentInstruction() {
	switch vm.currInstruction.opcode {
	case Add:
		// add the input and store it inside the output
		vm.memory[vm.currInstruction.params[2]] = vm.getParamValue(0) + vm.getParamValue(1)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Multiply:
		// multiply the input and store it inside the output
		vm.memory[vm.currInstruction.params[2]] = vm.getParamValue(0) * vm.getParamValue(1)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Input:
		// store the input of the program
		vm.memory[vm.currInstruction.params[0]] = vm.input
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Output:
		// store the instruction output into the VM output and print it to the screen
		vm.output = vm.getParamValue(0)
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
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
		if vm.getParamValue(0) < vm.getParamValue(1) {
			vm.memory[vm.currInstruction.params[2]] = 1
		} else {
			vm.memory[vm.currInstruction.params[2]] = 0
		}
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	case Equals:
		if vm.getParamValue(0) == vm.getParamValue(1) {
			vm.memory[vm.currInstruction.params[2]] = 1
		} else {
			vm.memory[vm.currInstruction.params[2]] = 0
		}
		// increment the instruction pointer and memory
		vm.instructionPointer += vm.currInstruction.length
	default:
		panic("Unknown opcode encountered!")
	}
}
