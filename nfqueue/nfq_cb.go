package nfqueue

import (
    "unsafe"
)

import "C"

/* Cast argument to Queue* before calling the real callback
   Notes:
     - export cannot be done in the same file (nfqueue.go) else it
       fails to build (multiple definitions of C functions)
       See https://github.com/golang/go/issues/3497
       See https://github.com/golang/go/wiki/cgo
     - this cast is caused by the fact that cgo does not support
       exporting structs
       See https://github.com/golang/go/wiki/cgo
*/
//export GoCallbackWrapper
func GoCallbackWrapper(ptr_q *unsafe.Pointer, id uint32, data *unsafe.Pointer, payload_len int) int {
    q := (*Queue)(unsafe.Pointer(ptr_q))
    payload := C.GoBytes(unsafe.Pointer(data), C.int(payload_len))
    verdict := q.cb(id,payload)
    q.SetVerdict(id,verdict)
    return 0
}


