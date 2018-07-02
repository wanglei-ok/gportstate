package portstate

import (
	"unsafe"
	"net"
	"fmt"
	"encoding/binary"
	"unicode/utf16"
)

/*
#include <stdlib.h>
 */
import 	"C"

type memBuffer struct {
	ptr  unsafe.Pointer
	size C.size_t
}

func alloc(sz uint32) *memBuffer {
	return &memBuffer{
		ptr:  C.malloc(C.size_t(sz)),
		size: C.size_t(sz),
	}
}

func (mb *memBuffer) resize(newSize C.size_t) {
	mb.ptr = C.realloc(mb.ptr, newSize)
	mb.size = newSize
}

func (mb *memBuffer) free() {
	C.free(mb.ptr)
}

func inet_aton(ip string) (ip_int uint32) {
	ip_byte := net.ParseIP(ip).To4()
	for i := 0; i < len(ip_byte); i++ {
		ip_int |= uint32(ip_byte[i])
		if i < 3 {
			ip_int <<= 8
		}
	}
	return
}

func inet_ntoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(ip>>24), byte(ip>>16), byte(ip>>8),
		byte(ip))
}

func ntohl(i uint32) uint32 {
	return binary.BigEndian.Uint32((*(*[4]byte)(unsafe.Pointer(&i)))[:])
}

func htonl(i uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return *(*uint32)(unsafe.Pointer(&b[0]))
}

func ntohs(i uint16) uint16 {
	return binary.BigEndian.Uint16((*(*[2]byte)(unsafe.Pointer(&i)))[:])
}
func htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}

func stringFromUnicode16(s *uint16) string {
	if s == nil {
		return ""
	}
	buffer := []uint16{}
	ptr := uintptr(unsafe.Pointer(s))
	for true {
		ch := *(*uint16)(unsafe.Pointer(ptr))
		if ch == 0 {
			break
		}
		buffer = append(buffer, ch)
		ptr += unsafe.Sizeof(ch)
	}
	return string(utf16.Decode(buffer))
}
