package main

import "fmt"

func main() {

	nums := []int{1, 2, 4, 5, 6}

	for i, num := range nums {
		fmt.Println(i, num)
	}

}
