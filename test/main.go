package main

import "fmt"

func main() {
    c := add([8]byte{1, 2, 3, 4, 5, 6, 7, 8}, [8]byte{1, 2, 3, 4, 5, 6, 7, 8})
    for i := 0; i < 8; i++ {
        fmt.Print(c[i])
    }
}
