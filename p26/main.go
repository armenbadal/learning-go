package main

import "fmt"

func f(yield func(x int) bool) {
	i := 0
	sum := 0
	for yield(sum) {
		sum += i
		i += 1
	}
}

func g(yield func(p, v int) bool) {
	i := 0
	sum := 0
calculate:
	i += 1
	sum += i
	if yield(i, sum) {
		goto calculate
	}
}

func main() {
	// 6. For a function f, the iteration proceeds by calling f
	// with a new, synthesized yield function as its argument.
	// If yield is called before f returns, the arguments to
	// yield become the iteration values for executing the loop
	// body once. After each successive loop iteration, yield
	// returns true and may be called again to continue the loop.
	// As long as the loop body does not terminate, the "range"
	// clause will continue to generate iteration values this way
	// for each yield call until f returns. If the loop body
	// terminates (such as by a break statement), yield returns
	// false and must not be called again.
	for i := range f {
		if i > 150 {
			break
		}
		fmt.Println(i)
	}

	fmt.Println("Second")
	for p, v := range g {
		if v > 150 {
			break
		}
		fmt.Printf("%d: %d\n", p, v)
	}
}
