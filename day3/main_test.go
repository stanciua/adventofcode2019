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

	newGrid := [][]rune{
		{'.', '.', '.', '.'},
		{'.', '1', '2', '.'},
		{'.', '3', '4', '.'},
		{'.', '.', '.', '.'},
	}

	if !reflect.DeepEqual(newGrid, resizeGrid(grid)) {
		t.Errorf("Resize grid not working!\n %v \n %v", grid, newGrid)
	}
}
