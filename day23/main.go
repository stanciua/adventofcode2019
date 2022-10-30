package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Computer struct {
	vm            *VM
	queue         []Packet
	received      Received
	sent          Sent
	currSent      Packet
	currRecv      Packet
	inputCounter  int64
	outputCounter int64
}

type Nat struct {
	lastPacket Packet
	seenY      map[int64]int64
}

type Received struct {
	x bool
	y bool
}

type Sent struct {
	dest bool
	x    bool
	y    bool
}

type Packet struct {
	dest int64
	x    int64
	y    int64
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
	And
	Or
	Not
	ProgramStop = 99
)

type ParameterMode int64

const (
	Positional ParameterMode = iota
	Immediate
	Relative
)

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

func part1(input []int64) int64 {
	computers := make([]Computer, 0)
	for i := 0; i < 50; i++ {
		comp := Computer{vm: NewVM()}
		comp.vm.loadProgram(input)
		comp.boot(int64(i))
		computers = append(computers, comp)
	}

	done, output := false, int64(0)
	for {
		done, output = communicate(computers, nil)
		if done {
			break
		}
	}

	return output
}

func receive(computers []Computer, addr int) {
	if len(computers[addr].queue) == 0 {
		computers[addr].vm.input = []int64{-1}
		computers[addr].inputCounter++
	} else {
		if !computers[addr].received.x {
			computers[addr].currRecv = computers[addr].queue[0]
			computers[addr].vm.input = []int64{computers[addr].currRecv.x}
			computers[addr].received.x = true
		} else if !computers[addr].received.y {
			computers[addr].vm.input = []int64{computers[addr].currRecv.y}
			computers[addr].received.y = true
		}
		if computers[addr].received.x && computers[addr].received.y {
			computers[addr].queue = computers[addr].queue[1:]
			computers[addr].received.x = false
			computers[addr].received.y = false
		}
	}
}
func send(computers []Computer, addr int, nat *Nat) (bool, int64) {
	done, output := false, int64(0)

	computers[addr].outputCounter++
	computers[addr].vm.outputReady = false
	if !computers[addr].sent.dest {
		computers[addr].currSent.dest = computers[addr].vm.output
		computers[addr].sent.dest = true
	} else if !computers[addr].sent.x {
		computers[addr].currSent.x = computers[addr].vm.output
		computers[addr].sent.x = true
	} else if !computers[addr].sent.y {
		computers[addr].currSent.y = computers[addr].vm.output
		computers[addr].sent.y = true
	}

	// send the packet if ready and then clear all the flags
	if computers[addr].sent.dest && computers[addr].sent.x && computers[addr].sent.y {
		if computers[addr].currSent.dest == 255 && computers[addr].sent.y {
			if nat == nil {
				done = true
				output = computers[addr].currSent.y
				return done, output
			} else {
				nat.lastPacket = computers[addr].currSent
				nat.lastPacket.dest = 0
			}
		} else {
			computers[computers[addr].currSent.dest].queue = append(computers[computers[addr].currSent.dest].queue, computers[addr].currSent)
		}

		computers[addr].sent.dest = false
		computers[addr].sent.x = false
		computers[addr].sent.y = false
	}

	return done, output
}

func communicate(computers []Computer, nat *Nat) (bool, int64) {
	done, output := false, int64(0)
	networkIdle := true
	for addr := range computers {
		computers[addr].vm.currInstruction = computers[addr].vm.decodeCurrentInstruction()
		if computers[addr].vm.currInstruction.opcode == Input {
			receive(computers, addr)
		}

		computers[addr].vm.executeCurrentInstruction()

		if computers[addr].vm.outputReady {
			if done, output = send(computers, addr, nat); done {
				return done, output
			}
		}

		//  check if all computers are in idle state: difference between input and output
		//  counters is greater than 240 for this input
		networkIdle = networkIdle && (computers[addr].inputCounter-computers[addr].outputCounter) > 240
	}

	// if network is idle
	if networkIdle {
		// reset the counters
		for i := 0; i < 50; i++ {
			computers[i].inputCounter = 0
			computers[i].outputCounter = 0
		}
		// check to see if we have seen Y value twice in a row
		if times, ok := nat.seenY[nat.lastPacket.y]; !ok {
			nat.seenY[nat.lastPacket.y] = 1
		} else {
			nat.seenY[nat.lastPacket.y] = times + 1
			done = true
			output = nat.lastPacket.y
		}

		// resume activity if Y value is not seen twice
		computers[0].queue = append(computers[0].queue, nat.lastPacket)
	}

	return done, output
}

func part2(input []int64) int64 {
	computers := make([]Computer, 0)
	for i := 0; i < 50; i++ {
		comp := Computer{vm: NewVM()}
		comp.vm.loadProgram(input)
		comp.boot(int64(i))
		computers = append(computers, comp)
	}

	seenY := make(map[int64]int64)
	nat := Nat{Packet{0, 0, 0}, seenY}

	done, output := false, int64(0)
	for {
		done, output = communicate(computers, &nat)
		if done {
			break
		}
	}

	return output
}

func (comp *Computer) boot(address int64) {
	done := false
	for {
		comp.vm.currInstruction = comp.vm.decodeCurrentInstruction()
		if comp.vm.currInstruction.opcode == Input {
			comp.vm.input = []int64{address}
			done = true

		}

		comp.vm.executeCurrentInstruction()

		if done {
			break
		}
	}
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
