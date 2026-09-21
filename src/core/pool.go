package core

import "sync"

var bytePool = sync.Pool{New: func() any {
	bs := []byte(nil)
	return &bs
}}

func allocBytes(size int) []byte {
	bs := *bytePool.Get().(*[]byte)
	if cap(bs) < size {
		bs = make([]byte, size)
	}
	return bs[:size]
}

func freeBytes(bs []byte) {
	bs = bs[:0]
	bytePool.Put(&bs)
}
