package main

import (
	"testing"
)

func TestPart2(t *testing.T) {
	if part2(445555, 445555) != 1 {
		t.Errorf("Incorrect algorithm!")
	}
	if part2(223344, 223344) != 1 {
		t.Errorf("Incorrect algorithm!")
	}
	if part2(136999, 136999) != 0 {
		t.Errorf("Incorrect algorithm!")
	}
}
