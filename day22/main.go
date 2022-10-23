package main

// Part 2 solved thanks to this Modulo Arithmetic tutorial:
// https://codeforces.com/blog/entry/72593

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
)

const (
	Cut       int = 0
	Increment int = 1
	NewStack  int = 2
)

type Technique struct {
	t int
	n int
}

// Linear Congruence Functions f(x) = ax + b mod m, is represented as
type Lcf struct {
	a *big.Int
	b *big.Int
	m *big.Int
}

const (
	DECK_SIZE_1 int64 = 10007
	DECK_SIZE_2 int64 = 119315717514047
	NO_SHUFFLES int64 = 101741582076661
)

func composeLcf(lcf0, lcf1 Lcf) Lcf {
	lcf0.a = lcf0.a.Mul(lcf0.a, lcf1.a)
	lcf0.a = lcf0.a.Mod(lcf0.a, lcf0.m)

	lcf0.b = lcf0.b.Mul(lcf0.b, lcf1.a)
	lcf0.b = lcf0.b.Add(lcf0.b, lcf1.b)
	lcf0.b = lcf0.b.Mod(lcf0.b, lcf0.m)

	return Lcf{lcf0.a, lcf0.b, lcf0.m}
}

// We are solving this congruence: f(x) = ax + b mod m
func solveLinearCongruence(lcf Lcf, x *big.Int) int64 {
	lcf.a = lcf.a.Mul(lcf.a, x)
	lcf.a = lcf.a.Add(lcf.a, lcf.b)
	lcf.a = lcf.a.Mod(lcf.a, lcf.m)

	return lcf.a.Int64()
}

func part1(techniques []Technique) int64 {
	m := big.NewInt(DECK_SIZE_1)
	lcfn := getLcf(techniques[0], m)

	for i := 1; i < len(techniques); i++ {
		lcfi := getLcf(techniques[i], m)
		lcfn = composeLcf(lcfn, lcfi)
	}

	return solveLinearCongruence(lcfn, big.NewInt(2019))
}

func getLcf(t Technique, m *big.Int) Lcf {
	a, b := int64(0), int64(0)
	if t.t == Cut {
		a, b = 1, int64(-t.n)
	} else if t.t == Increment {
		a, b = int64(t.n), 0
	} else {
		a, b = -1, -1
	}
	return Lcf{big.NewInt(a), big.NewInt(b), m}
}

// We are solving this linear inverted congruence: F^-k(x) = (x - b) * a^-1 mod m
// which represents the inverse of F^k(x) = ax + b mod m
func solveLinearCongruenceInverted(lcf Lcf, x *big.Int) int64 {
	inv := pow_mod(new(big.Int).Set(lcf.a), lcf.m, lcf.m.Int64()-2)
	lcf.b = lcf.b.Sub(x, lcf.b)
	lcf.b = lcf.b.Mod(lcf.b, lcf.m)
	lcf.b = lcf.b.Mul(lcf.b, inv)
	lcf.b = lcf.b.Mod(lcf.b, lcf.m)

	return lcf.b.Int64()
}

func part2(techniques []Technique) int64 {
	m := big.NewInt(DECK_SIZE_2)
	n := int64(NO_SHUFFLES)
	lcfn := getLcf(techniques[0], m)

	for i := 1; i < len(techniques); i++ {
		lcfi := getLcf(techniques[i], m)
		lcfn = composeLcf(lcfn, lcfi)
	}

	lcfFinal := pow_compose(lcfn, n)

	return solveLinearCongruenceInverted(lcfFinal, big.NewInt(2020))
}

func pow_mod(x, m *big.Int, k int64) *big.Int {
	y := big.NewInt(int64(1))
	for k > 0 {
		if k%2 != 0 {
			y = y.Mul(y, x)
			y = y.Mod(y, m)
		}
		k = k / 2
		x = x.Mul(x, x)
		x = x.Mod(x, m)
	}
	return y
}

func pow_compose(lcf Lcf, k int64) Lcf {
	lcff := Lcf{new(big.Int).Set(lcf.a), new(big.Int).Set(lcf.b), lcf.m}
	lcfg := Lcf{big.NewInt(1), big.NewInt(0), lcf.m}

	for k > 0 {
		if k%2 != 0 {
			lcfg = composeLcf(lcfg, lcff)
		}

		k = k / 2

		lcffCopy := Lcf{new(big.Int).Set(lcff.a), new(big.Int).Set(lcff.b), lcff.m}
		lcff = composeLcf(lcff, lcffCopy)
	}

	return lcfg
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
}
