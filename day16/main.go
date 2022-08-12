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

var PATTERN = [...]int{0, 1, 0, -1}

func getPatternForDigit(digit int, maxLen int) []int {
	uniquePattern := make([]int, 0)
	for _, d := range PATTERN {
		for i := 1; i <= digit; i++ {
			uniquePattern = append(uniquePattern, d)
		}
	}

	repeatedPattern := uniquePattern[:]
	idx := 0
	// need to take into account that we need extra element because we are going to strip
	// it from the begining
	for len(repeatedPattern) < maxLen+1 {
		repeatedPattern = append(repeatedPattern, uniquePattern[idx%len(uniquePattern)])
		idx++
	}
	return repeatedPattern[1:]
}

func precomputePatternsForEachDigit(maxLen int) map[int][]int {
	precomputedPatterns := make(map[int][]int)
	for i := 1; i <= maxLen; i++ {
		precomputedPatterns[i] = getPatternForDigit(i, maxLen)
	}

	return precomputedPatterns
}

func getInputSignalDigits(inputSignal string) []int {
	inputDigits := make([]int, 0)
	for _, d := range inputSignal {
		inputDigits = append(inputDigits, int(d-'0'))
	}

	return inputDigits
}

func updateInputSignal(inputDigits []int, precomputedPatterns map[int][]int) {
	for i := 0; i < len(inputDigits); i++ {
		sum := 0
		for idx, d := range inputDigits {
			pattern := precomputedPatterns[i+1]
			sum += d * pattern[idx]
		}

		inputDigits[i] = int(math.Abs(float64(sum % 10)))
	}
}

func updateInputSignalOptimized(inputDigits []int, initialInputDigits []int, iteration int) {
	sum := 0
	for _, d := range inputDigits {
		sum += d
	}
	prevDigit := inputDigits[0]
	inputDigits[0] = int(math.Abs(float64(sum % 10)))

	for idx, d := range inputDigits[1:] {
		sum -= prevDigit
		prevDigit = d
		inputDigits[idx+1] = int(math.Abs(float64(sum % 10)))
	}
}

func part1(inputSignal string) string {
	inputDigits := getInputSignalDigits(inputSignal)
	precomputedPatterns := precomputePatternsForEachDigit(len(inputDigits))
	for i := 0; i < 100; i++ {
		updateInputSignal(inputDigits, precomputedPatterns)
	}

	output := make([]string, len(inputDigits[:8]))
	for _, d := range inputDigits[:8] {
		output = append(output, strconv.Itoa(d))
	}

	return strings.Join(output, "")
}

func part2(inputSignal string) string {
	// get the offset first
	offset, err := strconv.ParseInt(inputSignal[:7], 10, 0)
	if err != nil {
		fmt.Println("invalid int conversion for string: ", string(inputSignal[:7]))
		return ""
	}
	// repeat actual input 10.000 times
	input := inputSignal
	for i := 0; i < 9999; i++ {
		input += inputSignal
	}

	inputDigits := getInputSignalDigits(input[offset:])
	initialInputDigits := getInputSignalDigits(inputSignal)
	for i := 0; i < 100; i++ {
		updateInputSignalOptimized(inputDigits, initialInputDigits, i)
	}

	output := make([]string, len(inputDigits[:8]))
	for _, d := range inputDigits[:8] {
		output = append(output, strconv.Itoa(d))
	}

	return strings.Join(output, "")
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

	inputSignal := inputs[0]
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(inputSignal))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(inputSignal))
}
