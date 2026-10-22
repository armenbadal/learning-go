package main

type Counter struct {
	n int
}

func (c *Counter) Inc() {
	c.n++
}

type Incrementer interface {
	Inc()
}

func main() {
	var c Counter
	c.Inc()
	println(c.n)

	var a Incrementer = &c
	a.Inc()
	println(c.n)
}
