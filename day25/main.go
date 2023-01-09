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
	North int = 1
	South int = 2
	East  int = 3
	West  int = 4
)

type Pos struct {
	x int
	y int
}

type Move struct {
	pos Pos
	dir int
}

type Room struct {
	m        Move
	doors    [4]int
	item     string
	security bool
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

func (d *Droid) parseOutput(m Move, output string) Room {
	room := Room{m, [4]int{}, "", false}
	lines := strings.Split(output, "\n")
	doorsFlag := false
	doors := make([]int, 0)
	itemsFlag := false
	securityFlag := false
	for _, l := range lines {
		if strings.HasPrefix(l, "Doors here lead:") {
			doorsFlag = true
			continue
		} else if strings.HasPrefix(l, "Items here:") {
			itemsFlag = true
			continue
		} else if strings.HasPrefix(l, "== Security Checkpoint ==") {
			securityFlag = true
			break
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
			item := strings.TrimSpace(l)
			item = strings.Trim(item, "- ")
			room.item = item
			if item != "infinite loop" &&
				item != "escape pod" &&
				item != "molten lava" &&
				item != "giant electromagnet" &&
				item != "photons" {
				d.input([]rune("take " + item + "\n"))
			}
			itemsFlag = false
			continue
		}
	}

	copy(room.doors[:], doors)
	room.security = securityFlag

	return room
}

func adjacentEdges(s *Room) []Move {
	moves := make([]Move, 0)
	for _, d := range s.doors {
		if d == 0 {
			continue
		}

		switch d {
		case North:
			moves = append(moves, Move{Pos{s.m.pos.x, s.m.pos.y - 1}, d})
		case South:
			moves = append(moves, Move{Pos{s.m.pos.x, s.m.pos.y + 1}, d})
		case West:
			moves = append(moves, Move{Pos{s.m.pos.x - 1, s.m.pos.y}, d})
		case East:
			moves = append(moves, Move{Pos{s.m.pos.x + 1, s.m.pos.y}, d})
		}
	}

	return moves
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

func (d *Droid) move(dir int) {
	var m string
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

func (d *Droid) searchEnv(m Move, explored map[[32]rune]bool, neighbors map[Pos]map[Pos]bool, progress []Pos, checkpoint bool) {
	Q := make([]Move, 0)
	Q = append(Q, m)

	for len(Q) > 0 {
		v := Q[len(Q)-1]
		Q = Q[:len(Q)-1]

		// we need to backtrack to v position from where we are with the droid
		from := progress[len(progress)-1]
		for _, ok := neighbors[from][v.pos]; !ok; _, ok = neighbors[from][v.pos] {
			to := progress[len(progress)-2]
			dir := findDir(from, to)
			d.move(dir)
			progress = progress[:len(progress)-1]
			from = progress[len(progress)-1]
		}

		d.move(v.dir)

		progress = append(progress, v.pos)
		output := d.output()
		var r [32]rune
		copy(r[:], []rune(getRoomId(output)))
		explored[r] = true
		s := d.parseOutput(v, output)
		doors := adjacentEdges(&s)
		n := make(map[Pos]bool)
		for _, d := range doors {
			n[d.pos] = true
		}
		neighbors[v.pos] = n
		if !s.security {
			for _, w := range doors {
				if !d.addRoomIfNotExplored(v, w, explored) {
					Q = append(Q, w)
				}
			}
		} else {
			if checkpoint {
				return
			}
		}
	}
}

func getCombinations(items []string, n, r int) [][]string {
	combinations := make([][]string, 0)
	combination := make([]string, r)
	combinationsUtil(items, &combinations, combination, 0, n-1, 0, r)
	return combinations
}

func combinationsUtil(items []string, combinations *[][]string, combination []string, start, end, index, r int) {
	if index == r {
		*combinations = append(*combinations, append([]string(nil), combination...))
		return
	}

	for i := start; i <= end && end-i+1 >= r-index; i++ {
		combination[index] = items[i]
		combinationsUtil(items, combinations, combination, i+1, end, index+1, r)
	}
}

func (d *Droid) addRoomIfNotExplored(curr, next Move, explored map[[32]rune]bool) bool {
	visited := true

	// move to next room to see if it's visited
	d.move(next.dir)

	output := d.output()
	lines := strings.Split(output, "\n")
	var room string
	for _, line := range lines {
		if line != "" {
			room = strings.TrimSpace(line)
			break
		}
	}

	var e [32]rune
	copy(e[:], []rune(room))
	if _, ok := explored[e]; !ok {
		explored[e] = true
		visited = false
	}

	// backtrack
	p := findDir(next.pos, curr.pos)
	d.move(p)

	return visited
}

func reveseDir(dir int) int {
	r := North
	switch dir {
	case North:
		r = South
	case South:
		r = North
	case West:
		r = East
	case East:
		r = West
	}

	return r
}

func part1(input []int64) int64 {
	// Note: the code is specific to my inputs, it's not generic for other type of inputs
	//       as it take too much to handle every bad input generically.
	comp := Droid{vm: NewVM()}
	comp.vm.loadProgram(input)

	// search for all the items
	explored := make(map[[32]rune]bool)
	neighbors := make(map[Pos]map[Pos]bool)
	output := comp.output()
	var e [32]rune
	copy(e[:], []rune(getRoomId(output)))
	explored[e] = true
	neighbors[Pos{0, 0}] = map[Pos]bool{{0, 1}: true}
	progress := make([]Pos, 0)
	progress = append(progress, Pos{0, 0})
	comp.searchEnv(Move{Pos{0, 1}, 2}, explored, neighbors, progress, false)

	// search for the checkpoint now
	explored = make(map[[32]rune]bool)
	neighbors = make(map[Pos]map[Pos]bool)
	output = comp.output()
	var ne [32]rune
	copy(ne[:], []rune(getRoomId(output)))
	explored[e] = true
	neighbors[Pos{0, 0}] = map[Pos]bool{{0, 1}: true}
	progress = make([]Pos, 0)
	progress = append(progress, Pos{0, 0})

	comp.searchEnv(Move{Pos{0, 1}, 4}, explored, neighbors, progress, true)
	comp.input([]rune("inv\n"))
	items := itemsFromInventory(comp.output())

	return comp.findCode(items)
}

func (d *Droid) findCode(items []string) int64 {
	// brute force all the items combinations to find the code
	for r := 1; r <= 8; r++ {
		combinations := getCombinations(items, 8, r)
		for c := range combinations {
			// drop all items
			for _, item := range items {
				d.input([]rune("drop " + item + "\n"))
			}
			for _, i := range combinations[c] {
				d.input([]rune("take " + i + "\n"))
			}
			d.move(West)
			output := d.output()
			if strings.Contains(output, "Oh, hello!") {
				for _, line := range strings.Split(output, "\n") {
					if !strings.Contains(line, "Oh, hello!") {
						continue
					}

					for _, w := range strings.Split(line, " ") {
						w = strings.TrimSpace(w)
						code, err := strconv.ParseInt(w, 0, 64)
						if err == nil {
							return code
						}
					}
				}
			}
		}
	}

	return -1
}

func itemsFromInventory(output string) []string {
	items := make([]string, 0)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "- ") {
			continue
		}

		item := strings.Trim(line, "- ")
		item = strings.TrimSpace(item)

		items = append(items, item)
	}

	return items
}

func getRoomId(output string) string {
	lines := strings.Split(output, "\n")
	var room string
	for _, line := range lines {
		if line != "" {
			room = strings.TrimSpace(line)
			break
		}
	}

	return room
}

func (d *Droid) output() string {
	var output strings.Builder

	for {
		d.vm.currInstruction = d.vm.decodeCurrentInstruction()
		d.vm.executeCurrentInstruction()
		if d.vm.outputReady {
			output.WriteRune(rune(d.vm.output))
			d.vm.outputReady = false
			if strings.HasSuffix(output.String(), "Command?\n") ||
				strings.Contains(output.String(), "airlock.") {
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
}
