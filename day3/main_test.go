package main

import (
	"reflect"
	"testing"
)

func TestResizeGrid(t *testing.T) {
	grid := [][]rune{
		{'1', '2'},
		{'3', '4'},
	}

	expectedGrid := [][]rune{
		{' ', ' ', ' ', ' '},
		{' ', '1', '2', ' '},
		{' ', '3', '4', ' '},
		{' ', ' ', ' ', ' '},
	}

	actualGrid := resizeGrid(grid)
	if !reflect.DeepEqual(expectedGrid, actualGrid) {
		t.Errorf("Resize grid not working!\n %v \n %v", expectedGrid, actualGrid)
	}
}

func TestMap(t *testing.T) {
	positions := make(map[Coordinate]void)
	c := Coordinate{x: 2, y: 3}

	positions[c] = member
	if _, present := positions[c]; !present {
		t.Errorf("Element is not in map!")
	}
}
