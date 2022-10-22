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
	technique int
	n         int
}

const MAX_SIZE int64 = 10007

func part1(techniques []Technique) int64 {
	fa, fb := getAB(techniques[0])
	m := big.NewInt(MAX_SIZE)

	for i := 1; i < len(techniques); i++ {
		a, b := getAB(techniques[i])
		// fa, fb = (fa*a)%m, (fb*a+b)%m
		fa = fa.Mul(fa, a)
		fa = fa.Mod(fa, m)

		fb = fb.Mul(fb, a)
		fb = fb.Add(fb, b)
		fb = fb.Mod(fb, m)
	}

	fmt.Println(fa, fb)
	fa = fa.Mul(fa, big.NewInt(2019))
	fa = fa.Add(fa, fb)
	fa = fa.Mod(fa, m)

	return fa.Int64()
}

func getAB(t Technique) (*big.Int, *big.Int) {
	a, b := int64(0), int64(0)
	if t.technique == Cut {
		a, b = 1, int64(-t.n)
	} else if t.technique == Increment {
		a, b = int64(t.n), 0
	} else {
		a, b = -1, -1
	}
	return big.NewInt(a), big.NewInt(b)
}

// 49283089762689
func part2(techniques []Technique) int64 {
	t1 := big.NewInt(1)
	t2 := t1.Add(t1, t1)
	fmt.Println(t1.Int64(), t2.Int64())
	fa, fb := getAB(techniques[0])
	m := big.NewInt(119315717514047)
	n := int64(101741582076661)

	for i := 1; i < len(techniques); i++ {
		a, b := getAB(techniques[i])
		fa = fa.Mul(fa, a)
		fa = fa.Mod(fa, m)

		fb = fb.Mul(fb, a)
		fb = fb.Add(fb, b)
		fb = fb.Mod(fb, m)
	}

	a, b := geometric_series(fa, fb, m, n)
	fmt.Println("geometric_series: ", a, b)
	fa, fb = pow_compose(fa, fb, m, n)
	fmt.Println("pow_compose: ", fa, fb)
	
	inv := pow_mod(fa, m, m.Int64()-2)
	fb = fb.Sub(big.NewInt(2020), fb)
	fb = fb.Mod(fb, m)
	fb = fb.Mul(fb, inv)
	fb = fb.Mod(fb, m)

	return fb.Int64()
}

// 𝐹𝑘(𝑥)=𝑎𝑘𝑥+𝑏(1−𝑎𝑘)1−𝑎  mod 𝑚

func geometric_series(fa, fb, m *big.Int, k int64) (*big.Int, *big.Int) {
	aK := pow_mod(fa, m, k)
	Fa := big.NewInt(aK.Int64())
	Fa = Fa.Sub(big.NewInt(1), aK)
	Fa = Fa.Mul(Fa, fb)
	Fa = Fa.Mod(Fa, m)

	a := big.NewInt(fa.Int64())
	a = a.Sub(big.NewInt(1) , fa)
	inv := pow_mod(a, m, m.Int64()-2)
	inv = inv.Mod(inv, m)

	Fa = Fa.Mul(Fa, inv)
	Fa = Fa.Mod(Fa, m)


	return aK, Fa
}

func pow_mod(x, m *big.Int, n int64) *big.Int {
	y := big.NewInt(int64(1))
	X := big.NewInt(x.Int64())
	for n > 0 {
		if n%2 != 0 {
			y = y.Mul(y, X)
			y = y.Mod(y, m)
		}
		n = n / 2
		X = X.Mul(X, X)
		X = X.Mod(X, m)
	}
	return y
}

func pow_compose(fa, fb, m *big.Int, k int64) (*big.Int, *big.Int) {
	Fa := big.NewInt(fa.Int64())
	Fb := big.NewInt(fb.Int64())
	ga := big.NewInt(1)
	gb := big.NewInt(0)

	// (𝑎,𝑏) ;(𝑐,𝑑)=(𝑎𝑐 mod 𝑚,𝑏𝑐+𝑑  mod 𝑚)
	for k > 0 {
		if k%2 != 0 {
			ga = ga.Mul(ga, Fa)
			ga = ga.Mod(ga, m)
			gb = gb.Mul(gb, Fa)
			gb = gb.Mod(gb, m)
			gb = gb.Add(gb, Fb)
			gb = gb.Mod(gb, m)
		}

		k = k / 2

		a, b := big.NewInt(Fa.Int64()), big.NewInt(Fb.Int64())
		Fa = Fa.Mul(a, a)
		Fa = Fa.Mod(Fa, m)

		Fb = Fb.Mul(b, a)
		Fb = Fb.Mod(Fb, m)
		Fb = Fb.Add(Fb, b)
		Fb = Fb.Mod(Fb, m)
	}

	return ga, gb
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
