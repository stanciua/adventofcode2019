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
		startPos, present := twoAdjacentDigitsStartingPos(digits)
		if !present || !doDigitsIncreaseOrStayTheSame(digits) {
			continue
		}

		// see if we can find the double digits in another larger group, get its
		// start position, its length and a flag if it's part of a larger group
		pos, groupLen, found := moreThenTwoAdjacentDigitsPosition(digits, startPos)

		// now if we are done with larger group check, we need to look for other
		// valid double digits values that may be valid
		foundLeft := checkSubsliceForTwoAdjacentDigits(digits[0:pos])
		foundRight := checkSubsliceForTwoAdjacentDigits(digits[startPos+groupLen : len(digits)])
		if foundLeft || foundRight || !found {
			noOfPasswords++
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
	newStartPos := startPos
	newEndPos := startPos
	for i, j := startPos, startPos-1; i > 0 && j >= 0; i, j = i-1, j-1 {
		if digits[i] == digits[j] {
			newStartPos = j
		} else {
			break
		}
	}

	// look on the right side
	length := len(digits)
	for i, j := startPos, startPos+1; i < length && j < length; i, j = i+1, j+1 {
		if digits[i] == digits[j] {
			newEndPos = j
		} else {
			break
		}
	}

	length = newEndPos - newStartPos + 1
	return newStartPos, length, length != 2
}

func checkSubsliceForTwoAdjacentDigits(digits []int) bool {
	count := 0
	length := len(digits)
	if length >= 2 {
		for i, j := 0, 1; i < length && j < length; i, j = i+1, j+1 {
			if digits[i] == digits[j] {
				count++
			}
		}

	}
	return count == 1
}
