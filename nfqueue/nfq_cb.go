package nfqueue

import (
    "unsafe"
)

import "C"

/*
Cast argument to Queue* before calling the real callback

Notes:
  - export cannot be done in the same file (nfqueue.go) else it
    fails to build (multiple definitions of C functions)
    See https://github.com/golang/go/issues/3497
    See https://github.com/golang/go/wiki/cgo
  - this cast is caused by the fact that cgo does not support
    exporting structs
    See https://github.com/golang/go/wiki/cgo

This function must _nerver_ be called directly.
*/
/*
BUG(GoCallbackWrapper): The return value from the Go callback is used as a
verdict. This works, and avoids packets without verdict to be queued, but
prevents using out-of-order replies.
*/
//export GoCallbackWrapper
func GoCallbackWrapper(ptr_q *unsafe.Pointer, id uint32, data *unsafe.Pointer, payload_len int) int {
    q := (*Queue)(unsafe.Pointer(ptr_q))
    payload := C.GoBytes(unsafe.Pointer(data), C.int(payload_len))
    verdict := q.cb(id,payload)
    q.SetVerdict(id,verdict)
    return 0
}


