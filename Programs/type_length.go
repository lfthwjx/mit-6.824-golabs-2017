package main

import (
	"fmt"
	// "syscall"
	"unsafe"
)

func main() {
	a := 12
	fmt.Println("length of a: ", unsafe.Sizeof(a))
	var b int = 12
	fmt.Println("length of b(int): ", unsafe.Sizeof(b))
	var c int8 = 12
	fmt.Println("length of c(int8): ", unsafe.Sizeof(c))
	d := 99.999
	fmt.Println("length of d(float): ", unsafe.Sizeof(d))
	var dd int16 = 12
	fmt.Println("length of dd(int16):", unsafe.Sizeof(dd))
	var e int32 = 32
	fmt.Println("length of e(int32): ", unsafe.Sizeof(e))
	var f int64 = 12
	fmt.Println("length of f(int64):", unsafe.Sizeof(f))
	t1 := "\"hello\""
	t2 := `"hello"`
	t3 := "\u6B22\u8FCE"
	fmt.Println(t1, t2, t3)

}
