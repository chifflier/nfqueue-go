// +build libnfqueue1

package nfqueue

// This file contains code specific to versions >= 1.0 of libnetfilter_queue

/*
#include <stdio.h>
#include <stdint.h>
#include <arpa/inet.h>
#include <linux/netfilter.h>
#include <libnetfilter_queue/libnetfilter_queue.h>
*/
import "C"

import (
    "log"
    "unsafe"
)

// SetVerdictMark issues a verdict for a packet, but a mark can be set
//
// Every queued packet _must_ have a verdict specified by userspace.
func (q *Queue) SetVerdictMark(id uint32, verdict int, mark uint32) error {
    log.Printf("Setting verdict for packet %d: %d mark %lx\n",id,verdict,mark)
    C.nfq_set_verdict2(
        q.c_qh,
        C.u_int32_t(id),
        C.u_int32_t(verdict),
        C.u_int32_t(mark),
        0,nil)
    return nil
}

// SetVerdictMarkModified issues a verdict for a packet, but replaces the
// packet with the provided one, and a mark can be set.
//
// Every queued packet _must_ have a verdict specified by userspace.
func (q *Queue) SetVerdictMarkModified(id uint32, verdict int, mark uint32, data []byte) error {
    log.Printf("Setting verdict for NEW packet %d: %d mark %lx\n",id,verdict,mark)
    C.nfq_set_verdict2(
        q.c_qh,
        C.u_int32_t(id),
        C.u_int32_t(verdict),
        C.u_int32_t(mark),
        C.u_int32_t(len(data)),
        (*C.uchar)(unsafe.Pointer(&data[0])),
    )
    return nil
}
