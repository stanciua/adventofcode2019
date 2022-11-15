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
	Antenna            int = 0
	EscapePod          int = 1
	InfiniteLoop       int = 2
	SpaceHeater        int = 3
	MoltenLava         int = 4
	FixedPoint         int = 5
	EasterEgg          int = 6
	FestiveHat         int = 7
	GiantElectromagnet int = 8
)

const (
	North int = 1
	South int = 2
	East  int = 3
	West  int = 4
)

type Pos struct {
	x int
	y int
}

type Surrounding struct {
	pos                    Pos
	doors                  [4]int
	items                  [16]bool
	inventory              [16]bool
	locked                 bool
	festiveHatFlag         bool
	security               bool
	giantElectromagnetFlag bool
}

type Droid struct {
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

func parseOutput(pos Pos, output string, parents map[Pos]Pos) Surrounding {
	fmt.Println(pos, parents[pos], dirStr(findDir(parents[pos], pos)), output)
	surrounding := Surrounding{pos, [4]int{}, [16]bool{}, [16]bool{}, false,false, false, false}
	lines := strings.Split(output, "\n")
	doorsFlag := false
	doors := make([]int, 0)
	itemsFlag := false
	items := make([]int, 0)
	lockedFlag := false
	securityFlag := false
	inventoryFlag := false
	giantElectromagnetFlag := false
	festiveHatFlag := false
	inventory := make([]int, 0)
	for _, l := range lines {
		if strings.HasPrefix(l, "Doors here lead:") {
			doorsFlag = true
			continue
		} else if strings.HasPrefix(l, "Items here:") {
			itemsFlag = true
			continue
		} else if strings.HasPrefix(l, "You can't go that way.") {
			lockedFlag = true
			break
		} else if strings.HasPrefix(l, "== Security Checkpoint ==") {
			securityFlag = true
			break
		} else if strings.HasPrefix(l, "Items in your inventory:") {
			inventoryFlag = true
			continue
		} else if strings.HasPrefix(l, "The giant electromagnet is stuck to you.  You can't move!!") {
			giantElectromagnetFlag = true
			break
		} else if strings.HasPrefix(l, "They're a little twisty and starting to look all alike.") {
			festiveHatFlag = true
			continue
		}

		if doorsFlag {
			if strings.HasPrefix(l, "- north") {
				doors = append(doors, North)
			} else if strings.HasPrefix(l, "- south") {
				doors = append(doors, South)
			} else if strings.HasPrefix(l, "- west") {
				doors = append(doors, West)
			} else if strings.HasPrefix(l, "- east") {
				doors = append(doors, East)
			} else {
				doorsFlag = false
				continue
			}
		}

		if itemsFlag {
			if strings.HasPrefix(l, "- infinite loop") {
				surrounding.items[InfiniteLoop] = true
			} else if strings.HasPrefix(l, "- escape pod") {
				surrounding.items[EscapePod] = true
			} else if strings.HasPrefix(l, "- antenna") {
				surrounding.items[Antenna] = true
			} else if strings.HasPrefix(l, "- space heater") {
				surrounding.items[SpaceHeater] = true
			} else if strings.HasPrefix(l, "- molten lava") {
				surrounding.items[MoltenLava] = true
			} else if strings.HasPrefix(l, "- fixed point") {
				surrounding.items[FixedPoint] = true
			} else if strings.HasPrefix(l, "- easter egg") {
				surrounding.items[EasterEgg] = true
			} else if strings.HasPrefix(l, "- festive hat") {
				if festiveHatFlag {
					festiveHatFlag = false
				}
				surrounding.items[FestiveHat] = true
			} else if strings.HasPrefix(l, "- giant electromagnet") {
				surrounding.items[GiantElectromagnet] = true
			} else if strings.HasPrefix(l, "- ") {
				fmt.Println(l)
				items = append(items, 0)
			} else {
				itemsFlag = false
				continue
			}
		}

		if inventoryFlag {
			if strings.HasPrefix(l, "- infinite loop") {
				surrounding.inventory[InfiniteLoop] = true
			} else if strings.HasPrefix(l, "- escape pod") {
				surrounding.inventory[EscapePod] = true
			} else if strings.HasPrefix(l, "- antenna") {
				surrounding.inventory[Antenna] = true
			} else if strings.HasPrefix(l, "- space heater") {
				surrounding.inventory[SpaceHeater] = true
			} else if strings.HasPrefix(l, "- molten lava") {
				surrounding.inventory[MoltenLava] = true
			} else if strings.HasPrefix(l, "- fixed point") {
				surrounding.inventory[FixedPoint] = true
			} else if strings.HasPrefix(l, "- easter egg") {
				surrounding.inventory[EasterEgg] = true
			} else if strings.HasPrefix(l, "- festive hat") {
				surrounding.inventory[FestiveHat] = true
			} else if strings.HasPrefix(l, "- giant electromagnet") {
				surrounding.inventory[GiantElectromagnet] = true
			} else if strings.HasPrefix(l, "- ") {
				fmt.Println(l)
				inventory = append(inventory, 0)
			} else {
				inventoryFlag = false
				continue
			}
		}
	}

	copy(surrounding.doors[:], doors)
	surrounding.locked = lockedFlag
	surrounding.festiveHatFlag = festiveHatFlag
	surrounding.security = securityFlag
	surrounding.giantElectromagnetFlag = giantElectromagnetFlag

	return surrounding
}

func adjacentEdges(s *Surrounding) []Pos {
	positions := make([]Pos, 0)
	for _, d := range s.doors {
		if d == 0 {
			continue
		}

		switch d {
		case North:
			positions = append(positions, Pos{s.pos.x, s.pos.y - 1})
		case South:
			positions = append(positions, Pos{s.pos.x, s.pos.y + 1})
		case West:
			positions = append(positions, Pos{s.pos.x - 1, s.pos.y})
		case East:
			positions = append(positions, Pos{s.pos.x + 1, s.pos.y})
		}
	}

	return positions
}

func findDir(from Pos, to Pos) int {
	// North
	d := 0
	pos := Pos{from.x, from.y - 1}
	if to == pos {
		d = North
	}
	// South
	pos = Pos{from.x, from.y + 1}
	if to == pos {
		d = South
	}
	// West
	pos = Pos{from.x - 1, from.y}
	if to == pos {
		d = West
	}
	// East
	pos = Pos{from.x + 1, from.y}
	if to == pos {
		d = East
	}

	return d
}

func dirStr(dir int) string {
	var ds string
	switch dir {
	case North:
		ds = "north"
	case South:
		ds = "south"
	case West:
		ds = "west"
	case East:
		ds = "east"
	default:
		ds = "UNKNOWN"
	}

	return ds
}

func (d *Droid) move(pos Pos, parents map[Pos]Pos) {
	var m string
	dir := findDir(parents[pos], pos)
	switch dir {
	case North:
		m = "north\n"
	case South:
		m = "south\n"
	case West:
		m = "west\n"
	case East:
		m = "east\n"
	default:
		m = ""
	}

	if len(m) > 0 {
		d.input([]rune(m))
	}
}

// func (d *Droid) unlock(p Pos, surroundings map[Pos]Surrounding, pastItems map[Pos]map[int]bool) {
// 	// check to see if the door is locked
// 	for _, i := range [...]int{Antenna, EscapePod, InfiniteLoop, SpaceHeater, MoltenLava} {
// 		if pastItems[p][i] {
// 			continue
// 		}
// 		s := surroundings[p]
// 		d.take(s)
// 		d.move(p, parents)
// 		s = parseOutput(p, d.output())
// 		if s.locked == false {
// 			pastItems[p][i] = true
// 			break
// 		}
// 	}
// }

func (d *Droid) searchEnv(start Pos, parents map[Pos]Pos) int {
	Q := make([]Pos, 0)
	explored := make(map[Pos]bool)
	explored[start] = true
	surroundings := make(map[Pos]Surrounding)
	pastItems := make(map[Pos]map[int]bool)
	Q = append(Q, start)

	for len(Q) > 0 {
		fmt.Println(Q)
		v := Q[0]
		Q = Q[1:]

		if _, ok := pastItems[v]; !ok {
			items := make(map[int]bool)
			pastItems[v] = items
		}
		d.move(v, parents)
		s := parseOutput(v, d.output(), parents)
		// take items if present
		d.take(s, parents)
		d.move(v, parents)
		surroundings[v] = s
		if s.giantElectromagnetFlag {
			d.input([]rune("drop giant electromagnet\n"))
			d.move(v, parents)
			s = parseOutput(v, d.output(), parents)
		}

		// if s.locked {
		// 	d.unlock(v, surroundings, pastItems)
		// 	d.move(v)
		// 	s = parseOutput(v, d.output())
		// }

		// if s.security {
		// 	d.input([]rune("inv\n"))
		// 	s = parseOutput(v, d.output())
		// 	for _, i := range s.inventory {
		// 		if i == 0 {
		// 			continue
		// 		}
		// 		var command string
		// 		switch i {
		// 		case Antenna:
		// 			command = "drop antenna\n"
		// 		case EscapePod:
		// 			command = "drop escape pod\n"
		// 		case InfiniteLoop:
		// 			command = "drop infinite loop\n"
		// 		case SpaceHeater:
		// 			command = "drop space heater\n"
		// 		case MoltenLava:
		// 			command = "drop molten lava\n"
		// 		}
		//
		// 		d.input([]rune(command))
		// 		s = parseOutput(v, d.output())
		// 	}
		//
		// 	d.move(v, parents)
		// 	s = parseOutput(v, d.output())
		// }

		surroundings[v] = s

		// if s.locked {
		// 	explored[s.pos] = true
		// 	Q = append(Q, s.pos)
		// }
		for _, w := range adjacentEdges(&s) {
			parents[w] = v
			// check to see if this door has been locked
			// sw, ok1 := surroundings[w]
			// _, ok2 := explored[w]

			// if (ok1 && sw.locked) || !ok2 {
			if s.items[MoltenLava] ||
			s.festiveHatFlag {
				// w == (Pos{0, 0}) {
				continue
			}
			// if !ok2 {
			// explored[w] = true
			Q = append(Q, w)
			fmt.Println("******queued: ", w, "******")
			// } else {
			// fmt.Println("******explored: ", w, "******")
			// }
		}
		fmt.Println("_____________________________________")
	}

	return -1
}

func (d *Droid) take(s Surrounding, parents map[Pos]Pos) {
	for idx, i := range s.items {
		if !i {
			continue
		}

		var command string
		switch idx {
		case Antenna:
			command = "take antenna\n"
		case SpaceHeater:
			command = "take space heater\n"
		case EasterEgg:
			command = "take easter egg\n"
		case FixedPoint:
			command = "take fixed point\n"
		case FestiveHat:
			command = "take festive hat\n"
		// case GiantElectromagnet:
		// 	command = "take giant electromagnet\n"
		default:
			continue
		}

		d.input([]rune(command))
		_ = parseOutput(s.pos, d.output(), parents)
	}
}

func part1(input []int64) int64 {
	comp := Droid{vm: NewVM()}
	comp.vm.loadProgram(input)

	parents := make(map[Pos]Pos)
	return int64(comp.searchEnv(Pos{0, 0}, parents))
}

func (d *Droid) output() string {
	var output strings.Builder

	for {
		d.vm.currInstruction = d.vm.decodeCurrentInstruction()
		d.vm.executeCurrentInstruction()
		if d.vm.outputReady {
			output.WriteRune(rune(d.vm.output))
			d.vm.outputReady = false
			if strings.HasSuffix(output.String(), "Command?\n") {
				break
			}
		}
	}

	return output.String()
}

func (d *Droid) input(command []rune) {
	for len(command) > 0 {
		d.vm.currInstruction = d.vm.decodeCurrentInstruction()
		if d.vm.currInstruction.opcode == Input {
			d.vm.input = []int64{int64(command[0])}
			command = command[1:]
		}
		d.vm.executeCurrentInstruction()
	}
}

func part2(input []int64) int64 {
	comp := Droid{vm: NewVM()}
	comp.vm.loadProgram(input)
	return -1
}

func (comp *Droid) boot(address int64) {
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
