package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

func requiredFuelForMass(mass int) int {
	return mass/3 - 2
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

	// part 1 solution
	fmt.Println(part1(inputs))

	// part 2 solution
	fmt.Println(part2(inputs))
}

func part1(inputs []string) int {
	totalFuel := 0
	for _, input := range inputs {
		if mass, err := strconv.Atoi(input); err == nil {
			fuel := requiredFuelForMass(mass)
			totalFuel += fuel
		}
	}
	return totalFuel
}

func part2(inputs []string) int {
	totalFuel := 0
	for _, input := range inputs {
		if mass, err := strconv.Atoi(input); err == nil {
			var fuels []int
			realFuel := calculateTotalFuelRequired(requiredFuelForMass(mass), fuels)
			totalFuel += realFuel
		}
	}
	return totalFuel
}

func calculateTotalFuelRequired(fuel int, fuels []int) int {
	if fuel <= 0 {
		totalFuel := 0
		for _, elem := range fuels {
			totalFuel += elem
		}
		return totalFuel
	}

	fuels = append(fuels, fuel)
	return calculateTotalFuelRequired(requiredFuelForMass(fuel), fuels)
}
