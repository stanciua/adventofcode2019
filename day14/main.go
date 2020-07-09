package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Chemical struct {
	name     string
	quantity int64
}

type Reaction struct {
	input  []Chemical
	output Chemical
}

func part1(reactions map[string]*Reaction) int64 {
	produced := make(map[string]int64)
	produced["FUEL"] = 1
	reduceReaction(produced, reactions)
	return produced["ORE"]
}

func reduceReaction(produced map[string]int64, reactions map[string]*Reaction) {
	for {
		changed := false
		for name, quantity := range produced {
			if quantity > 0 {
				if reaction, ok := reactions[name]; ok {
					changed = true
					factor := (quantity + reaction.output.quantity - 1) / reaction.output.quantity
					produced[name] -= factor * reaction.output.quantity
					for _, input := range reaction.input {
						produced[input.name] += factor * input.quantity
					}
				}
			}
		}
		if !changed {
			return
		}
	}
}

func getFuelNumber(adder int64, initialFuel int64, fuelReaction *Reaction, reactions map[string]*Reaction) int64 {
	for {
		produced := make(map[string]int64)
		produced["FUEL"] = initialFuel
		reactions["FUEL"] = &Reaction{
			input:  append([]Chemical(nil), fuelReaction.input...),
			output: fuelReaction.output,
		}
		reactions["FUEL"].output.quantity *= initialFuel
		for i := range reactions["FUEL"].input {
			reactions["FUEL"].input[i].quantity *= initialFuel
		}

		reduceReaction(produced, reactions)

		if produced["ORE"] <= 1_000_000_000_000 {
			initialFuel += adder
		} else {
			break
		}
	}

	return initialFuel
}
func part2(reactions map[string]*Reaction) int64 {
	c := int64(1)
	fuelReaction := Reaction{
		input:  append([]Chemical(nil), reactions["FUEL"].input...),
		output: reactions["FUEL"].output,
	}

	c = getFuelNumber(10_000, 1, &fuelReaction, reactions)
	c -= 10_000
	c = getFuelNumber(1, c, &fuelReaction, reactions)

	return c - 1
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

	reactions := make(map[string]*Reaction)
	for _, p := range inputs {
		reaction := strings.Split(p, "=>")
		// output chemical processing
		var outputChemical Chemical
		nameQuantity := strings.Split(strings.Trim(reaction[1], " "), " ")
		if val, err := strconv.ParseInt(nameQuantity[0], 10, 64); err != nil {
			panic(err)
		} else {
			outputChemical = Chemical{
				name:     strings.Trim(nameQuantity[1], " "),
				quantity: val,
			}
		}

		// input chemical processing
		input := strings.Trim(reaction[0], " ")
		inputChemicals := strings.Split(input, ",")
		var chemicals []Chemical
		for _, c := range inputChemicals {
			nameQuantity := strings.Split(strings.Trim(c, " "), " ")
			if val, err := strconv.ParseInt(nameQuantity[0], 10, 64); err != nil {
				panic(err)
			} else {
				chemical := Chemical{
					name:     strings.Trim(nameQuantity[1], " "),
					quantity: val,
				}
				chemicals = append(chemicals, chemical)
			}
		}

		chemicalReaction := &Reaction{
			input:  chemicals,
			output: outputChemical,
		}

		reactions[chemicalReaction.output.name] = chemicalReaction
	}
	// part 1 solution
	fmt.Println("The result to 1st part is: ", part1(reactions))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ", part2(reactions))
}
