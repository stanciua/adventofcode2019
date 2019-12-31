package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
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

	// we have just one line, with elements separated by commas
	if len(inputs) != 1 {
		panic("The input should be only one line long!")
	}

	var rawImage []int
	for _, c := range inputs[0] {
		rawImage = append(rawImage, int(c)-int('0'))
	}

	fmt.Println("The result to 1st part is: ", part1(rawImage))

	// part 2 solution
	fmt.Println("The result to 2nd part is: ")
	displayPassword(part2(rawImage))
}

type Layer struct {
	data   [][]int
	height int
	width  int
}

func newLayer(height, width int) Layer {
	data := make([][]int, height)
	for i := 0; i < height; i++ {
		data[i] = make([]int, width)
	}

	return Layer{data: data, height: height, width: width}
}

func parseRawImage(rawImage []int) []Layer {
	var layers []Layer
	lgth := len(rawImage)
	idx := 0
	for idx < lgth {
		layer := newLayer(6, 25)
		for i := 0; i < layer.height; i += 1 {
			for j := 0; j < layer.width; j += 1 {
				layer.data[i][j] = rawImage[idx+(i*layer.width+j)]
			}
		}
		idx += layer.height * layer.width
		layers = append(layers, layer)
	}

	return layers
}

func countLayerDigit(layer Layer, digit int) int {
	count := 0
	for i := 0; i < layer.height; i++ {
		for j := 0; j < layer.width; j++ {
			if layer.data[i][j] == digit {
				count++
			}
		}
	}
	return count
}
func part1(input []int) int {
	minZeroes := math.MaxInt32
	idx := 0
	layers := parseRawImage(input)
	for i, layer := range layers {
		zeroesCount := countLayerDigit(layer, 0)
		if zeroesCount < minZeroes {
			minZeroes = zeroesCount
			idx = i
		}
	}

	return countLayerDigit(layers[idx], 1) * countLayerDigit(layers[idx], 2)
}

func displayPassword(password Layer) {
	for i := 0; i < password.height; i++ {
		for j := 0; j < password.width; j++ {
			if password.data[i][j] == 1 {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func part2(input []int) Layer {
	layers := parseRawImage(input)
	password := newLayer(6, 25)
	for i := 0; i < password.height; i++ {
		for j := 0; j < password.width; j++ {
			for _, layer := range layers {
				if layer.data[i][j] == 1 {
					password.data[i][j] = 1
					break
				} else if layer.data[i][j] == 0 {
					password.data[i][j] = 0
					break
				} else {
					continue
				}
			}
		}
	}
	return password
}
