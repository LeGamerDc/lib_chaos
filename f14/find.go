package f14

import (
    "math/bits"
    "unsafe"
)

type d8 [2]uint16
type imm128 [16]byte

func find(a, b imm128) (c uint32)

func find2(a, b imm128) int {
    for i := 0; i < 16; i++ {
        if a[i] == b[i] {
            return i
        }
    }
    return -1
}

func find3(a, b imm128) int {
    mask := find(a, b)
    //return int(mask)
    return 15 - bits.LeadingZeros16((*(*d8)(unsafe.Pointer(&mask)))[0])
}
