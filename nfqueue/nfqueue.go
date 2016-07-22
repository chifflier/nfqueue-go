package nfqueue

// XXX we should use something like
// pkg-config --libs libnetfilter_queue

// #cgo pkg-config: libnetfilter_queue
/*
#include <stdio.h>
#include <stdint.h>
#include <arpa/inet.h>
#include <linux/netfilter.h>
#include <libnetfilter_queue/libnetfilter_queue.h>

extern int GoCallbackWrapper(void *data, int id, unsigned char*, int len);

int _process_loop(struct nfq_handle *h,
                  int fd,
                  int flags,
                  int max_count) {
        int rv;
        char buf[65535];
        int count;

        count = 0;

        while ((rv = recv(fd, buf, sizeof(buf), flags)) >= 0) {
                nfq_handle_packet(h, buf, rv);
                count++;
                if (max_count > 0 && count >= max_count) {
                        break;
                }
        }
        return count;
}

int c_nfq_cb(struct nfq_q_handle *qh,
             struct nfgenmsg *nfmsg,
             struct nfq_data *nfad, void *data) {
    int id = 0;
    struct nfqnl_msg_packet_hdr *ph;
    unsigned char *payload_data;
    int payload_len;

    ph = nfq_get_msg_packet_hdr(nfad);
    if (ph){
        id = ntohl(ph->packet_id);
    }

    if ((payload_len = nfq_get_payload(nfad, &payload_data)) < 0) {
        fprintf(stderr, "Couldn't get payload\n");
        return -1;
    }


    return GoCallbackWrapper(data, id, payload_data, payload_len);
}
*/
import "C"

import (
    "errors"
    "log"
    "unsafe"
)

var ErrNotInitialized = errors.New("nfqueue: queue not initialized")
var ErrOpenFailed = errors.New("nfqueue: open failed")
var ErrRuntime = errors.New("nfqueue: runtime error")

var NF_DROP = C.NF_DROP
var NF_ACCEPT = C.NF_ACCEPT
var NF_QUEUE = C.NF_QUEUE
var NF_REPEAT = C.NF_REPEAT
var NF_STOP = C.NF_STOP


type Callback func(uint32,[]byte) int

type Queue struct {
    c_h (*C.struct_nfq_handle)
    c_qh (*C.struct_nfq_q_handle)

    cb Callback
}

func (q *Queue) Init() error {
    log.Println("Opening queue")
    q.c_h = C.nfq_open()
    if (q.c_h == nil) {
        log.Println("nfq_open failed")
        return ErrOpenFailed
    }
    return nil
}

func (q *Queue) SetCallback(cb Callback) error {
    q.cb = cb
    return nil
}

func (q *Queue) Close() {
    if (q.c_h != nil) {
        log.Println("Closing queue")
        C.nfq_close(q.c_h)
        q.c_h = nil
    }
}

func (q *Queue) Bind(af_family int) error {
    if (q.c_h == nil) {
        return ErrNotInitialized
    }
    log.Println("Binding to selected family")
    /* Errors in nfq_bind_pf are non-fatal ...
     * This function just tells the kernel that nfnetlink_queue is
     * the chosen module to queue packets to userspace.
     */
    _ = C.nfq_bind_pf(q.c_h,C.u_int16_t(af_family))
    return nil
}

func (q *Queue) Unbind(af_family int) error {
    if (q.c_h == nil) {
        return ErrNotInitialized
    }
    log.Println("Unbinding to selected family")
    rc := C.nfq_unbind_pf(q.c_h,C.u_int16_t(af_family))
    if (rc < 0) {
        log.Println("nfq_unbind_pf failed")
        return ErrRuntime
    }
    return nil
}

// Set callback function
func (q *Queue) CreateQueue(queue_num int) error {
    if (q.c_h == nil) {
        return ErrNotInitialized
    }
    log.Println("Creating queue")
    q.c_qh = C.nfq_create_queue(q.c_h,C.u_int16_t(queue_num),(*C.nfq_callback)(C.c_nfq_cb),unsafe.Pointer(q))
    if (q.c_qh == nil) {
        log.Println("nfq_create_queue failed")
        return ErrRuntime
    }
    return nil
}

// Main loop
func (q *Queue) TryRun() error {
    if (q.c_h == nil) {
        return ErrNotInitialized
    }
    if (q.c_qh == nil) {
        return ErrNotInitialized
    }
    if (q.cb == nil) {
        return ErrNotInitialized
    }
    log.Println("Try Run")
    fd := C.nfq_fd(q.c_h)
    if (fd < 0) {
        log.Println("nfq_fd failed")
        return ErrRuntime
    }
    // XXX
    C.nfq_set_mode(q.c_qh,C.NFQNL_COPY_PACKET,0xffff)
    C._process_loop(q.c_h,fd,0,-1)
    return nil
}


func (q *Queue) SetVerdict(id uint32, verdict int) error {
    log.Printf("Setting verdict for packet %d: %d\n",id,verdict)
    C.nfq_set_verdict(q.c_qh,C.u_int32_t(id),C.u_int32_t(verdict),0,nil)
    return nil
}
