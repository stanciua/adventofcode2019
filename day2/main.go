package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const terminationResult int = 19690720

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
	// restore the "1202 program alarm"
	return executeProgramWithInputs(input, 12, 2)
}

func part2(input []int) int {
	for noun := range [100]int{} {
		for verb := range [100]int{} {
			inputCopy := append(input[:0:0], input...)
			if executeProgramWithInputs(inputCopy, noun, verb) == terminationResult {
				return 100*noun + verb
			}
		}
	}
	return -1
}

func executeProgramWithInputs(input []int, noun int, verb int) int {
	input[1] = noun
	input[2] = verb
	program := input[:]
	for program[0] != 99 {
		opcodeData := program[:4]
		executeOpcode(opcodeData, input)
		program = program[4:]
	}
	return input[0]
}

func executeOpcode(opcodeData []int, program []int) {
	opcode := opcodeData[0]
	input1 := opcodeData[1]
	input2 := opcodeData[2]
	output := opcodeData[3]
	switch opcode {
	case 1:
		// add the input and store it inside the output
		program[output] = program[input1] + program[input2]
	case 2:
		// multiply the input and store it inside the output
		program[output] = program[input1] * program[input2]
	default:
		panic("Unknown opcode encountered!")
	}
}
