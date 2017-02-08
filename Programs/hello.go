package main

import (
	"fmt"
	"golang.org/x/tour/pic"
	"golang.org/x/tour/wc"
	"math"
	"math/cmplx"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

func add(x, y int) int {
	return x + y
}

func swap(x, y string) (string, string) {
	return y, x
}

//naked return
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

var c, python, java bool
var i, j int = 1, 2
var (
	ToBe   bool       = false
	MaxInt uint64     = 1<<64 - 1
	z      complex128 = cmplx.Sqrt(-5 + 12i)
)

const (
	Big   = 1 << 100
	Small = Big >> 99
)

func needInt(x int) int {
	return x*10 + 1
}

func needFloat(x float64) float64 {
	return x * 0.1
}

func sqrt(x float64) string {
	if x < 0 {
		return sqrt(-x) + "i"
	}
	return fmt.Sprint(math.Sqrt(x))
}

func pow(x, n, lim float64) float64 {
	if v := math.Pow(x, n); v < lim {
		return v
	}
	return lim
}

func Sqrt(x float64) float64 {
	z := float64(2.)
	s := float64(0)

	for {
		z = z - (z*z-x)/(2*z)
		if math.Abs(s-z) < 1e-15 {
			break
		}
		s = z
	}
	return s
}

type Vertex struct {
	X int
	Y int
}

var (
	v1 = Vertex{1, 2}  // has type Vertex
	v2 = Vertex{X: 1}  // Y:0 is implicit
	v3 = Vertex{}      // X:0 and Y:0
	p  = &Vertex{1, 2} // has type *Vertex
)

// func printSlice(s []int) {
// 	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
// }

func printSlice(s string, x []int) {
	fmt.Printf("%s len=%d cap=%d %v\n",
		s, len(x), cap(x), x)
}

var pow2 = []int{1, 2, 4, 8, 16, 32, 64, 128}

func Pic(dx, dy int) [][]uint8 {
	ret := make([][]uint8, dy)
	for i := 0; i < dy; i++ {
		ret[i] = make([]uint8, dx)
		for j := 0; j < dx; j++ {
			ret[i][j] = uint8(i ^ j + (i+j)/2)
		}
	}
	return ret
}

type vertex struct {
	Lat, Long float64
}

var m = map[string]vertex{
	"Bell Labs": {40.68433, -74.39967},
	"Google":    {37.42202, -122.08408},
}

func WordCount(s string) map[string]int {
	m := make(map[string]int)
	for _, key := range strings.Fields(s) {
		m[key]++
	}
	return m
}

func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func fibonacci() func() int {
	a := 0
	b := 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

type dot struct {
	X, Y float64
}

//a method is just a function with a receiver argument.
func (d dot) Abs() float64 {
	return math.Sqrt(d.X*d.X + d.Y*d.Y)
}

type MyFloat float64

func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func (d *dot) Scale(f float64) {
	d.X = d.X * f
	d.Y = d.Y * f
}

func scale(d *dot, f float64) {
	d.X = d.X * f
	d.Y = d.Y * f
}

func abs(d dot) float64 {
	return math.Sqrt(d.X*d.X + d.Y*d.Y)
}

func main() {
	fff := MyFloat(-math.Sqrt2)
	fmt.Println(fff.Abs())

	dd := dot{3, 4}
	dd.Scale(10)
	fmt.Println(dd.Abs())
	dd = dot{3, 4}
	scale(&dd, 10)
	fmt.Println(abs(dd))

	fib := fibonacci()
	for i := 0; i < 10; i++ {
		fmt.Println(fib())
	}
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-i),
		)
	}

	hypot := func(x, y float64) float64 {
		return math.Sqrt(x*x + y*y)
	}
	fmt.Println(hypot(5, 12))

	fmt.Println(compute(hypot))
	fmt.Println(compute(math.Pow))

	wc.Test(WordCount)
	mm := make(map[string]int)
	mm["Answer"] = 42
	delete(mm, "Answer")
	value, within := mm["Answer"]
	fmt.Println("The value:", value, "Present?", within)
	pic.Show(Pic)
	//m = make(map[string]vertex)
	//m["Bell Labs"] = vertex{40.6843, -74.39967}
	fmt.Println(m["Bell Labs"])

	root := 2.0
	fmt.Println(Sqrt(root))
	pow3 := make([]int, 10)
	for i := range pow3 {
		pow3[i] = 1 << uint(i)
	}
	for _, value := range pow3 {
		fmt.Printf("%d\n", value)
	}

	for i, v := range pow2 {
		fmt.Printf("2**%d = %d\n", i, v)
	}
	// Create a tic-tac-toe board.
	board := [][]string{
		{"_", "_", "_"},
		{"_", "_", "_"},
		{"_", "_", "_"},
	}

	// The players take turns.
	board[0][0] = "X"
	board[2][2] = "O"
	board[1][2] = "X"
	board[1][0] = "O"
	board[0][2] = "X"

	for i := 0; i < len(board); i++ {
		fmt.Printf("%s\n", strings.Join(board[i], " "))
	}

	aaa := make([]int, 5)
	printSlice("a", aaa)

	b := make([]int, 0, 5)
	printSlice("b", b)

	c := b[:2]
	printSlice("c", c)

	d := c[2:5]
	printSlice("d", d)

	e := d[1:3]
	printSlice("e", e)

	ff := e[0:2]
	printSlice("f", ff)

	// sa := []int{2, 3, 5, 7, 11, 13}
	// printSlice(sa)
	// sa = sa[:0]
	// printSlice(sa)
	// sa = sa[:4]
	// printSlice(sa)

	q := []int{2, 3, 5, 7, 11, 13}
	fmt.Println(q)
	r := []bool{true, false, true, true, false, true}
	fmt.Println(r)

	s := []struct {
		i int
		b bool
	}{
		{2, true},
		{3, false},
		{5, true},
		{7, true},
		{11, false},
		{13, true},
	}
	fmt.Println(s)

	ss := []int{2, 3, 5, 7, 11, 13}

	ss = ss[1:4]
	fmt.Println(ss)

	ss = ss[:2]
	fmt.Println(ss)

	ss = ss[1:]
	fmt.Println(ss)

	names := [4]string{
		"John",
		"Paul",
		"George",
		"Ringo",
	}
	fmt.Println(names)

	aa := names[0:2]
	ba := names[1:3]
	fmt.Println(aa, ba)

	ba[0] = "XXX"
	fmt.Println(aa, ba)
	fmt.Println(names)

	fmt.Println(v1, p, v2, v3)
	ver := Vertex{1, 2}
	ver.X = 4
	fmt.Println(ver.X)
	fmt.Println(Vertex{1, 2})
	var a [2]string
	a[0] = "Hello"
	a[1] = "World"
	fmt.Println(a[0], a[1])
	fmt.Println(a)

	primes := [6]int{2, 3, 5, 7, 11, 13}
	fmt.Println(primes)
	var slice []int = primes[1:4]
	fmt.Println(slice)
	//v := Vertex{1, 2}
	// p := &ver
	// p.X = 1e9
	// fmt.Println(ver)

	// p2 := ver
	// p2.X = 3
	// fmt.Println(ver.X)
	// i, j := 42, 2701
	// p := &i
	// *p = 21
	// fmt.Println(*p)
	// fmt.Println(i)
	// fmt.Println(p)
	// p = &j
	// *p = *p / 37
	// fmt.Println(j)
	// fmt.Println("counting")

	// for i := 0; i < 10; i++ {
	// 	defer fmt.Println(i)
	// }

	fmt.Println("done")
	defer fmt.Println("dragon")
	fmt.Println("When's Saturday?")
	today := time.Now().Weekday()
	switch time.Saturday {
	case today + 0:
		fmt.Println("Today.")
	case today + 1:
		fmt.Println("Tomorrow.")
	case today + 2:
		fmt.Println("In two days.")
	default:
		fmt.Println("Too far away.")
	}

	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("Good morning!")
	case t.Hour() < 17:
		fmt.Println("Good afternoon.")
	default:
		fmt.Println("Good evening.")
	}

	fmt.Print("Go runs on ")
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("OS X.")
	case "linux":
		fmt.Println("Linux.")
	default:
		// freebsd, openbsd,
		// plan9, windows...
		fmt.Printf("%s.", os)
		fmt.Print("\n")
	}
	fmt.Println(Sqrt(2))
	fmt.Println(pow(3, 2, 10),
		pow(3, 3, 10),
	)
	fmt.Println(sqrt(2), sqrt(-4))
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)
	sum = 1
	for sum < 1000 {
		sum += sum
	}
	fmt.Println(sum)

	sum = 1
	for sum < 1000 {
		sum += sum
	}
	fmt.Println(sum)

	fmt.Println(needInt(Small))
	fmt.Println(needFloat(Small))
	fmt.Println(needFloat(Big))
	target := "World"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	fmt.Println("hello, jun", target)
	fmt.Println("Myfavorite number is", rand.Intn(10))
	fmt.Printf("Now you have %g problems.\n", math.Sqrt(7))
	fmt.Println(math.Pi) //notice that Capital letter beginning stands for exported name
	fmt.Println(add(42, 13))
	// a, b := swap("jun", "hello")
	// fmt.Println(a, b)
	// fmt.Println(split(17))
	//var i int
	// fmt.Println(i, c, python, java)
	// var c, python, java = true, false, "no!"
	// fmt.Println(i, j, c, python, java)

	//var i, j int = 1, 2
	// k := 3
	//c, python, java := true, false, true
	//fmt.Println(i, j, k, c, python, java)
	fmt.Printf("Type: %T Value: %v\n", ToBe, ToBe)
	fmt.Printf("Type: %T Value: %v\n", MaxInt, MaxInt)
	fmt.Printf("Type: %T Value: %v\n", z, z)

	var integer int
	var f float64
	var boo bool
	//var s string
	fmt.Printf("%v %v %v %q\n", integer, f, boo, s)

	var x, y int = 3, 4
	fl := math.Sqrt(float64(x*x + y*y))
	var z uint = uint(fl)
	fmt.Println(x, y, z)

}
