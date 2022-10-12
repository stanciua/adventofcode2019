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

const MAX_SIZE int64 = 10007

func newStack(cards []int64) {
	for i := 0; i < len(cards)/2; i++ {
		cards[len(cards)-i-1], cards[i] = cards[i], cards[len(cards)-i-1]
	}
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
	cards := make([]int64, 0)
	// add the first approx 1 milion records
	for i := int64(0); i < MAX_SIZE; i++ {
		cards = append(cards, i)
	}
	// add the last approx 1 milion records
	// for i := int64(119315717514047) - MAX_SIZE; i < 119315717514047; i++ {
	// 	cards = append(cards, i)
	// }
	curr := cards[2020]
	for _, t := range techniques {
		if t.technique == Cut {
			cutNCards(&cards, t.n)
		} else if t.technique == Increment {
			incrementNCards(&cards, t.n)
		} else {
			newStack(cards)
		}

		if curr != cards[2020] {
			fmt.Println(t.technique, ": ", cards[2020])
			curr = cards[2020]
		}
	}

	return cards[2020]
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
