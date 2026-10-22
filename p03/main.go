package main

import "fmt"

type MyError struct{}

func (*MyError) Error() string {
	return "error"
}

func f() error {
	var e *MyError = nil
	return e
}

func main() {
	err := f()

	fmt.Println(err == nil)
}
