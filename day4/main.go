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

	// we have just one line, with numbers separated by -
	if len(inputs) != 1 {
		panic("The input should be only one line long!")
	}

	rng := strings.Split(inputs[0], "-")
	var rmin, rmax int
	if rmin, err = strconv.Atoi(rng[0]); err != nil {
		panic(err)
	}
	if rmax, err = strconv.Atoi(rng[1]); err != nil {
		panic(err)
	}

	fmt.Println(findAdjacentDigitsPositions(3, []int{1, 1, 1, 4, 5, 6}))

	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(rmin, rmax))
	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(rmin, rmax))
}

func getNumberDigits(number int) []int {
	var digits []int

	digits = append(digits, number/100_000)
	number %= 100_000
	digits = append(digits, number/10_000)
	number %= 10_000
	digits = append(digits, number/1_000)
	number %= 1_000
	digits = append(digits, number/100)
	number %= 100
	digits = append(digits, number/10)
	number %= 10
	digits = append(digits, number)

	return digits
}

func twoAdjacentDigitsAreTheSame(digits []int) bool {
	return digits[0] == digits[1] || digits[1] == digits[2] || digits[2] == digits[3] || digits[3] == digits[4] || digits[4] == digits[5]
}

func doDigitsIncreaseOrStayTheSame(digits []int) bool {
	return digits[0] <= digits[1] && digits[1] <= digits[2] && digits[2] <= digits[3] && digits[3] <= digits[4] && digits[4] <= digits[5]
}

func part1(rmin, rmax int) int {
	noOfPasswords := 0
	for i := rmin; i <= rmax; i++ {
		digits := getNumberDigits(i)
		if twoAdjacentDigitsAreTheSame(digits) && doDigitsIncreaseOrStayTheSame(digits) {
			noOfPasswords++
		}
	}

	return noOfPasswords
}

// we need to bypass the following combinations
// aaaaaa
// aaaaab
// baaaaa
// aaaabc, b != c
// bcaaaa, b != c
// aaabcd, b != c, c != d
// baaacd, c != d
// bcaaad, b != c
// bcdaaa, b != c, c != d

func part2(rmin, rmax int) int {
	noOfPasswords := 0
	noOfPasswords++
	// 111122
	for i := rmin; i <= rmax; i++ {
		digits := getNumberDigits(i)

		if !twoAdjacentDigitsAreTheSame(digits) || !doDigitsIncreaseOrStayTheSame(digits) {
			continue
		}

		if (digits[0] == digits[1] && digits[1] == digits[2] && digits[2] == digits[3] && digits[3] == digits[4] && digits[4] == digits[5]) ||
			(digits[0] == digits[1] && digits[1] == digits[2] && digits[2] == digits[3] && digits[3] == digits[4]) ||
			(digits[1] == digits[2] && digits[2] == digits[3] && digits[3] == digits[4] && digits[4] == digits[5]) ||
			(digits[0] == digits[1] && digits[1] == digits[2] && digits[2] == digits[3] && digits[4] != digits[5]) ||
			(digits[0] != digits[1] && digits[2] == digits[3] && digits[3] == digits[4] && digits[4] == digits[5]) ||
			(digits[0] == digits[1] && digits[1] == digits[2] && digits[3] != digits[4] && digits[4] && digits[5]) ||
			(digits[1] == digits[2] && digits[2] == digits[3] && digits[4] != digits[5]) ||
			(digits[0] != digits[1] && digits[2] == digits[3] && digits[3] == digits[4]) ||
			(digits[0] != digits[1] && digits[1] != digits[2] && digits[2] == digits[3] && digits[3] == digits[4] && digits[4] == digits[5]) {
			continue
		}
	}
	return noOfPasswords
}

func findAdjacentDigitsPositions(noOfAdjacentDigits int, digits []int) []int {
	var positions []int
	for i, j := 0, 1; i < len(digits) && j < len(digits)-1; i, j = i+1, j+1 {
		if digits[i] == digits[j] {
			positions = append(positions, i)
		}
	}

	if len(positions) != noOfAdjacentDigits {
		return nil
	}

	return positions
}
