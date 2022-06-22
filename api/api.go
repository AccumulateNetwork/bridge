package api

import "fmt"

func zero(x int) {
	x = 0
}

func api() {
	x := 5
	zero(5)
	fmt.Println(x)

}
