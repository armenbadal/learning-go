package main

import "sort"

type cell struct {
	id string
}

type Cells []cell

func (c Cells) Len() int {
	return len(c)
}

func (c Cells) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Cells) Less(i, j int) bool {
	return c[i].id < c[j].id
}

func createAndSort() {
	set := []cell{
		cell{id: "a"},
		cell{id: "f"},
		cell{id: "d"},
		cell{id: "g"},
		cell{id: "h"},
		cell{id: "j"},
		cell{id: "i"},
		cell{id: "e"},
		cell{id: "c"},
		cell{id: "b"},
	}

	sort.Slice(set, func(i, j int) bool { return set[i].id < set[j].id })
	for _, e := range set {
		println(e.id)
	}
}

func createAndSort2() {
	set := []cell{
		cell{id: "a"},
		cell{id: "f"},
		cell{id: "d"},
		cell{id: "g"},
		cell{id: "h"},
		cell{id: "j"},
		cell{id: "i"},
		cell{id: "e"},
		cell{id: "c"},
		cell{id: "b"},
	}

	sort.Sort(Cells(set))
	for _, e := range set {
		println(e.id)
	}
}

func main() {
	createAndSort()
	println("---")
	createAndSort2()
}
