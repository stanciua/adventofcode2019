package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	HEIGHT int = 1200
	WIDTH  int = 1200
)

const (
	Stationary = '.'
	Pulled     = '#'
)

var DroneOutput = []rune{Stationary, Pulled}

type Drone struct {
	vm *VM
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

type BeamRow struct {
	begin int
	end   int
}

func (d *Drone) memoryChanges(program []int64) map[int]int64 {
	changes := make(map[int]int64)
	_ = d.deployDrone([]int64{0, 0})
	for i, v := range program {
		if d.vm.memory[i] != v {
			changes[i] = v
		}
	}

	return changes
}

func (d *Drone) resetDrone(changes map[int]int64) {
	for k, v := range changes {
		d.vm.memory[k] = v
	}

	d.vm.currInstruction = nil
	d.vm.instructionPointer = 0
}

func (d *Drone) buildBeam(input []int64, changes map[int]int64, height int, width int) ([][]rune, map[int]BeamRow, int, int) {
	view := make([][]rune, 0)
	for i := 0; i < height; i++ {
		line := make([]rune, 0)
		view = append(view, line)
		for j := 0; j < width; j++ {
			view[i] = append(view[i], Stationary)
		}
	}
	beamRows := make(map[int]BeamRow)
	begin := false
	startRow := 0
	endRow := 0
outer:
	for i := height - 1; i >= 0; i-- {
		fj, lj := 0, 0
		fSet := false
		for j := width - 1; j >= 0; j-- {
			d.resetDrone(changes)
			output := d.deployDrone([]int64{int64(i), int64(j)})
			if !fSet && output == 1 {
				fj = j
				fSet = true
			}
			if fSet && output == 0 {
				lj = j - 1
				fSet = false
				beamRow := BeamRow{lj, fj}
				beamRows[i] = beamRow
				view[i][j] = DroneOutput[output]
				// optimization for part 2
				if height > 50 {
					if !begin && fj-lj+1 >= 100 {
						endRow = i
						begin = true
					}
					if begin && fj-lj+1 < 100 {
						startRow = i
						break outer
					}
				}
				continue outer
			}
			view[i][j] = DroneOutput[output]
		}
	}

	return view, beamRows, startRow, endRow
}

func findClosestSquare(squareSize int, view [][]rune, beamRows map[int]BeamRow, startRow int, endRow int) int {
	output := 0
	for i := startRow; i <= endRow; i++ {
		b, ok := BeamRow{-1, -1}, false
		if b, ok = beamRows[i]; !ok {
			continue
		}
		y := i
		x := b.end
		// check top-right corner
		if !(view[y-1][x] == Stationary && view[y+1][x] == Pulled && view[y][x-1] == Pulled && view[y][x+1] == Stationary) {
			continue
		}

		// check bottom-left corner
		if view[i+squareSize-1][b.end+1-squareSize] != Pulled {
			continue
		}

		y = i + squareSize - 1
		x = b.end + 1 - squareSize
		if !(view[y-1][x] == Pulled && view[y+1][x] == Stationary && view[y][x-1] == Stationary && view[y][x+1] == Pulled) {
			continue
		}

		y = i
		x = b.end + 1 - squareSize
		output += y*10000 + x
		break
	}
	return output
}

func countPoints(view [][]rune) int {
	count := 0
	for i := 0; i < len(view); i++ {
		for j := 0; j < len(view[i]); j++ {
			if view[i][j] == Pulled {
				count++
			}
		}
	}

	return count
}

func part1(input []int64) int {
	d := Drone{vm: NewVM()}
	d.vm.loadProgram(input)
	changes := d.memoryChanges(input)
	view, _, _, _ := d.buildBeam(input, changes, 50, 50)
	return countPoints(view)
}

func part2(input []int64) int {
	d := Drone{vm: NewVM()}
	d.vm.loadProgram(input)
	changes := d.memoryChanges(input)
	view, beamRows, startRow, endRow := d.buildBeam(input, changes, HEIGHT, WIDTH)
	return findClosestSquare(100, view, beamRows, startRow, endRow)
}

func (robot *Drone) deployDrone(input []int64) int64 {
	output := int64(0)
	idx := 0

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
			output = robot.vm.output
			break
		}

		if robot.vm.hasFinished() {
			output = robot.vm.output
			break
		}
	}
	return output
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
