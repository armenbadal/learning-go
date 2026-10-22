package main

func Max[T interface {
	int | int64 | float32 | float64
}](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {

}
