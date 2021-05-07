package network

import "errors"

var ErrIsEmpty = errors.New("ring buffer is empty")

type RingBuffer struct {
	size    uint64
	rpos    uint64
	wpos    uint64
	buf     []byte
	isEmpty bool
	next    *NwBuffer
}

func New(size uint64) *RingBuffer {
	return &RingBuffer{
		buf:     make([]byte, size, size),
		size:    size,
		isEmpty: true,
	}
}

func NewWithData(data []byte) *RingBuffer {
	return &RingBuffer{
		buf:  data,
		size: uint64(len(data)),
	}
}

func (r *RingBuffer) WithData(data []byte) {
	r.rpos = 0
	r.wpos = 0
	r.isEmpty = false
	r.size = uint64(len(data))
	r.buf = data
}

func (r *RingBuffer) RetrieveAll() {
	r.rpos = 0
	r.wpos = 0
	r.isEmpty = true
}

func (r *RingBuffer) Retrieve(len uint64) {
	if r.isEmpty || len <= 0 {
		return
	}

	switch temp := len < r.freeReadSpace(); temp {
	case true:
		r.rpos = (r.rpos + len) % r.size
		if r.wpos == r.rpos {
			r.isEmpty = true
		}
	case false:
		r.RetrieveAll()
	}
}

func (r *RingBuffer) Read(p []byte) (n uint64, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.isEmpty {
		return 0, ErrIsEmpty
	}
	n = uint64(len(p))
	//wpos faster than rpos
	if r.wpos > r.rpos {
		if n > r.wpos-r.rpos {
			n = r.wpos - r.rpos
		}
		copy(p, r.buf[r.rpos:r.rpos+n])
		// move readPtr
		r.rpos = (r.rpos + n) % r.size
		if r.rpos == r.wpos {
			r.isEmpty = true
		}
		return
	}
	//r.wpos <= r.rpos
	//surplus read space
	if n > r.size-r.rpos+r.wpos {
		n = r.size - r.rpos + r.wpos
	}

	switch temp := r.rpos+n <= r.size; temp {
	case true:
		copy(p, r.buf[r.rpos:r.rpos+n])
	case false:
		// head
		copy(p, r.buf[r.rpos:r.size])
		// tail
		copy(p[r.size-r.rpos:], r.buf[0:n-r.size+r.rpos])
	}
	// move readPtr
	r.rpos = (r.rpos + n) % r.size
	if r.rpos == r.wpos {
		r.isEmpty = true
	}

	return
}

func (r *RingBuffer) makeSpace(len uint64) {
	newSize := r.size + len
	newBuf := make([]byte, newSize, newSize)
	oldLen := r.freeReadSpace()
	_, _ = r.Read(newBuf)

	r.wpos = oldLen
	r.rpos = 0
	r.size = newSize
	r.buf = newBuf
}

func (r *RingBuffer) Write(p []byte) (n uint64, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n = uint64(len(p))
	free := r.freeWritreSpace()
	if free < n {
		r.makeSpace(n - free)
	}
	switch temp := r.wpos >= r.rpos; temp {
	case true:
		if r.size-r.wpos >= n {
			copy(r.buf[r.wpos:], p)
			r.wpos += n
		} else {
			copy(r.buf[r.wpos:], p[:r.size-r.wpos])
			copy(r.buf[0:], p[r.size-r.wpos:])
			r.wpos += n - r.size
		}
	case false:
		copy(r.buf[r.wpos:], p)
		r.wpos += n
	}
	if r.wpos == r.size {
		r.wpos = 0
	}

	r.isEmpty = false

	return
}

//free_write_space
func (r *RingBuffer) freeWritreSpace() uint64 {
	if r.wpos == r.rpos {
		if r.isEmpty {
			return r.size
		}
		return 0
	}

	if r.wpos < r.rpos {
		return r.rpos - r.wpos
	}

	return r.size - r.wpos + r.rpos
}

//free_read_space
func (r *RingBuffer) freeReadSpace() uint64 {
	if r.wpos == r.rpos {
		if r.isEmpty {
			return 0
		}
		return r.size
	}

	if r.wpos > r.rpos {
		return r.wpos - r.rpos
	}

	return r.size - r.rpos + r.wpos
}

func (r *RingBuffer) Reset() {
	r.rpos = 0
	r.wpos = 0
	r.isEmpty = true
}

//
func (r *RingBuffer) PeekAll() (first []byte, end []byte) {
	if r.isEmpty {
		return
	}

	if r.wpos > r.rpos {
		first = r.buf[r.rpos:r.wpos]
		return
	}

	first = r.buf[r.rpos:r.size]
	end = r.buf[0:r.wpos]
	return
}
