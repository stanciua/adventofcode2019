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
		if _, present := twoAdjacentDigitsStartingPos(digits); present && doDigitsIncreaseOrStayTheSame(digits) {
			noOfPasswords++
		}
	}

	return noOfPasswords
}

func part2(rmin, rmax int) int {
	noOfPasswords := 0
	for i := rmin; i <= rmax; i++ {
		digits := getNumberDigits(i)
		if i == 136888 {
			fmt.Println(i)
		}
		startPos, present := twoAdjacentDigitsStartingPos(digits)
		if !present || !doDigitsIncreaseOrStayTheSame(digits) {
			continue
		}

		pos, groupLen, found := moreThenTwoAdjacentDigitsPosition(digits, startPos)

		// check the left side and right side
		if found {
			foundLeft := checkSubsliceForTwoAdjacentDigits(digits[0:pos])
			foundRight := checkSubsliceForTwoAdjacentDigits(digits[startPos+groupLen : len(digits)])
			if foundLeft || foundRight {
				noOfPasswords++
			}
		}

	}
	return noOfPasswords
}

func twoAdjacentDigitsStartingPos(digits []int) (int, bool) {
	if digits[0] == digits[1] {
		return 0, true
	} else if digits[1] == digits[2] {
		return 1, true
	} else if digits[2] == digits[3] {
		return 2, true
	} else if digits[3] == digits[4] {
		return 3, true
	} else if digits[4] == digits[5] {
		return 4, true
	}

	return -1, false
}

func moreThenTwoAdjacentDigitsPosition(digits []int, startPos int) (int, int, bool) {
	// look on the left side
	newStartPos := 0
	lenAdjacent := 2
	foundMoreThanTwoDigits := false
	for i := startPos - 1; i >= 0; i-- {
		if digits[startPos] == digits[i] {
			newStartPos = i
			lenAdjacent++
			foundMoreThanTwoDigits = true
		}
	}
	// look on the right side
	for i := startPos + 2; i < len(digits); i++ {
		if digits[startPos] == digits[i] {
			if !foundMoreThanTwoDigits {
				lenAdjacent++
				newStartPos = startPos
			} else {
				lenAdjacent++
			}
		}
	}

	return newStartPos, lenAdjacent, foundMoreThanTwoDigits
}

func checkSubsliceForTwoAdjacentDigits(digits []int) bool {
	count := 0
	if len(digits) >= 2 {
		i := 0
		j := i + 1
		for i < len(digits) && j < len(digits)-1 {
			if digits[i] == digits[j] {
				count++
			}
		}

	}
	return count == 1
}
