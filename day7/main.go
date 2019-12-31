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
	fmt.Println("The result to 1st part is: ", part1(program))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(program))
}

func part1(input []int) int {
	max := math.MinInt32
	output := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == j {
				continue
			}
			for k := 0; k < 5; k++ {
				if i == k || j == k {
					continue
				}
				for l := 0; l < 5; l++ {
					if i == l || j == l || k == l {
						continue
					}
					for m := 0; m < 5; m++ {
						if i == m || j == m || k == m || l == m {
							continue
						}
						// now start connecting the 5 amplifiers in serial
						// A
						vm := NewVM()
						vm.loadProgram(append(input[:0:0], input...))
						vm.input = []int{i, output}
						for vm.memory[vm.instructionPointer] != 99 {
							vm.currInstruction = vm.decodeCurrentInstruction()
							vm.executeCurrentInstruction()
						}
						output = vm.output
						// B
						vm = NewVM()
						vm.loadProgram(append(input[:0:0], input...))
						vm.input = []int{j, output}
						for vm.memory[vm.instructionPointer] != 99 {
							vm.currInstruction = vm.decodeCurrentInstruction()
							vm.executeCurrentInstruction()
						}
						output = vm.output
						// C
						vm = NewVM()
						vm.loadProgram(append(input[:0:0], input...))
						vm.input = []int{k, output}
						for vm.memory[vm.instructionPointer] != 99 {
							vm.currInstruction = vm.decodeCurrentInstruction()
							vm.executeCurrentInstruction()
						}
						output = vm.output
						// D
						vm = NewVM()
						vm.loadProgram(append(input[:0:0], input...))
						vm.input = []int{l, output}
						for vm.memory[vm.instructionPointer] != 99 {
							vm.currInstruction = vm.decodeCurrentInstruction()
							vm.executeCurrentInstruction()
						}
						output = vm.output
						// E
						vm = NewVM()
						vm.loadProgram(append(input[:0:0], input...))
						vm.input = []int{m, output}
						for vm.memory[vm.instructionPointer] != 99 {
							vm.currInstruction = vm.decodeCurrentInstruction()
							vm.executeCurrentInstruction()
						}
						output = 0
						if vm.output > max {
							max = vm.output
						}
					}
				}
			}
		}
	}
	return max
}

func part2(input []int) int {
	max := math.MinInt32
	output := 0
	for i := 5; i < 10; i++ {
		for j := 5; j < 10; j++ {
			if i == j {
				continue
			}
			for k := 5; k < 10; k++ {
				if i == k || j == k {
					continue
				}
				for l := 5; l < 10; l++ {
					if i == l || j == l || k == l {
						continue
					}
					for m := 5; m < 10; m++ {
						if i == m || j == m || k == m || l == m {
							continue
						}
						// now start connecting the 5 amplifiers in serial
						vmA := NewVM()
						vmA.loadProgram(append(input[:0:0], input...))
						vmA.input = append(vmA.input, i)
						vmB := NewVM()
						vmB.loadProgram(append(input[:0:0], input...))
						vmB.input = append(vmB.input, j)
						vmC := NewVM()
						vmC.loadProgram(append(input[:0:0], input...))
						vmC.input = append(vmC.input, k)
						vmD := NewVM()
						vmD.loadProgram(append(input[:0:0], input...))
						vmD.input = append(vmD.input, l)
						vmE := NewVM()
						vmE.loadProgram(append(input[:0:0], input...))
						vmE.input = append(vmE.input, m)
						output = 0
						for !vmE.hasFinished() {
							// A
							vmA.input = append(vmA.input, output)
							for vmA.outputReady == false && !vmA.hasFinished() {
								vmA.currInstruction = vmA.decodeCurrentInstruction()
								vmA.executeCurrentInstruction()
							}
							vmA.outputReady = false
							output = vmA.output
							// B
							vmB.input = append(vmB.input, output)
							for vmB.outputReady == false && !vmB.hasFinished() {
								vmB.currInstruction = vmB.decodeCurrentInstruction()
								vmB.executeCurrentInstruction()
							}
							vmB.outputReady = false
							output = vmB.output
							// C
							vmC.input = append(vmC.input, output)
							for vmC.outputReady == false && !vmC.hasFinished() {
								vmC.currInstruction = vmC.decodeCurrentInstruction()
								vmC.executeCurrentInstruction()
							}
							vmC.outputReady = false
							output = vmC.output
							// D
							vmD.input = append(vmD.input, output)
							for vmD.outputReady == false && !vmD.hasFinished() {
								vmD.currInstruction = vmD.decodeCurrentInstruction()
								vmD.executeCurrentInstruction()
							}
							vmD.outputReady = false
							output = vmD.output
							// E
							vmE.input = append(vmE.input, output)
							for vmE.outputReady == false && !vmE.hasFinished() {
								vmE.currInstruction = vmE.decodeCurrentInstruction()
								vmE.executeCurrentInstruction()
							}
							vmE.outputReady = false
							output = vmE.output
						}
						if output > max {
							max = output
						}
					}
				}
			}
		}
	}
	return max
}

type VM struct {
	memory             []int
	instructionPointer int
	input              []int
	output             int
	currInstruction    *Instruction
	outputReady        bool
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

func (vm *VM) hasFinished() bool {
	return vm.memory[vm.instructionPointer] == 99
}

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
	vm.outputReady = false
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
		vm.memory[vm.currInstruction.params[0]] = vm.input[0]
		vm.input = vm.input[1:]
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
