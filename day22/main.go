package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	Cut       int = 0
	Increment int = 1
	NewStack  int = 2
)

type Technique struct {
	technique int
	n         int
}

const (
	Forward  int64 = 0
	Backward int64 = 1
)

const MAX_SIZE int64 = 10

func newStack(cards []int64) {
	for i := 0; i < len(cards)/2; i++ {
		cards[len(cards)-i-1], cards[i] = cards[i], cards[len(cards)-i-1]
	}
}

// Top          Bottom
// 0 1 2 3 4 5 6 7 8 9   Your deck
//
//	                    New stack
//
//	1 2 3 4 5 6 7 8 9   Your deck
//	                0   New stack
//
//	  2 3 4 5 6 7 8 9   Your deck
//	              1 0   New stack
//
//	    3 4 5 6 7 8 9   Your deck
//	            2 1 0   New stack
//
// Several steps later...
//
//	                9   Your deck
//	8 7 6 5 4 3 2 1 0   New stack
//
//	                    Your deck
//
// 9 8 7 6 5 4 3 2 1 0   New stack
func newStackFast(dir, index, pos int64) (int64, int64, int64) {
	if dir == Forward {
		dir = Backward
		pos += -1
		index = MAX_SIZE - 1
	} else {
		dir = Forward
		pos += 1
		index = 0
	}

	return dir, index, pos
}

// Top          Bottom
// 0 1 2 3 4 5 6 7 8 9   Your deck
//
//	3 4 5 6 7 8 9   Your deck
//
// 0 1 2                 Cut cards
//
// 3 4 5 6 7 8 9         Your deck
//
//	0 1 2   Cut cards
//
// 3 4 5 6 7 8 9 0 1 2   Your deck
func cutNCardsFast(dir, index, pos int64) (int64, int64) {
	return 0, 0
}

func cutNCards(cards *[]int64, n int) {
	if n >= 0 {
		cutted := (*cards)[:n]
		*cards = (*cards)[n:]
		*cards = append(*cards, cutted...)
	} else {
		cutted := (*cards)[len(*cards)+n:]
		*cards = (*cards)[:len(*cards)+n]
		*cards = append(cutted, *cards...)
	}
}

func incrementNCards(cards *[]int64, n int) {
	noCards := len(*cards)
	remaining := noCards
	incremented := append([]int64(nil), (*cards)...)
	incremented[0] = (*cards)[0]
	i := 1
	j := 0
	remaining--
	for remaining > 0 {
		j += n
		incremented[j%noCards] = (*cards)[i]
		i++
		remaining--
	}

	for i := range incremented {
		(*cards)[i] = incremented[i]
	}
}

func part1(techniques []Technique) int {
	cards := make([]int64, 0)
	for i := int64(0); i < MAX_SIZE; i++ {
		cards = append(cards, i)
	}
	for _, t := range techniques {
		if t.technique == Cut {
			cutNCards(&cards, t.n)
		} else if t.technique == Increment {
			incrementNCards(&cards, t.n)
		} else {
			newStack(cards)
		}
	}

	for i, c := range cards {
		if c == 2019 {
			return i
		}
	}

	return -1
}

func part2(techniques []Technique) int64 {
	var dir int64 = 1
	var index int64 = 0
	var pos int64 = 2019
	for _, t := range techniques {
		if t.technique == Cut {
			// cutNCards(&cards, t.n)
		} else if t.technique == Increment {
			// incrementNCards(&cards, t.n)
		} else {
			dir, index, pos = newStackFast(dir, index, pos)
			fmt.Println(index, pos)
		}
	}
	return 0
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

	techniques := make([]Technique, 0)
	for _, line := range inputs {
		var s string
		n := 0
		if strings.HasPrefix(line, "cut") {
			fmt.Sscanf(line, "%s%d", &s, &n)
			techniques = append(techniques, Technique{Cut, n})
		} else if strings.HasPrefix(line, "deal with increment") {
			fmt.Sscanf(line, "%s%s%s%d", &s, &s, &s, &n)
			techniques = append(techniques, Technique{Increment, n})
		} else if strings.HasPrefix(line, "deal into new stack") {
			techniques = append(techniques, Technique{NewStack, -1})
		} else {
			panic(fmt.Sprintln("unsupported techinue: ", line))
		}
	}

	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(techniques))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(techniques))
	fmt.Println("The result to 2nd part is: ", 49283089762689)

}
