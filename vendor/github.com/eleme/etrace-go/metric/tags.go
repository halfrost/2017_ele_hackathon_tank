package metric

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"sync/atomic"

	"github.com/mailru/easyjson/jwriter"
)

type tag struct {
	lc       int32
	hash     hash.Hash64
	hashCode uint64
	values   map[string]string
	buf      [8]byte
}

func newTag() *tag {
	return &tag{
		hash:   fnv.New64(),
		values: make(map[string]string),
	}
}

func (t *tag) Add(key, val string) {
	if atomic.CompareAndSwapInt32(&t.lc, 0, 1) {
		t.values[key] = val
		code := t.hashPairs(key, val)
		t.hashCode = t.hashCode ^ code
		atomic.StoreInt32(&t.lc, 0)
	}
}

func (t *tag) HashCode() uint64 {
	return t.hashCode
}

func (t *tag) hashPairs(key, val string) uint64 {
	keyHash := t.hashString(key)
	valHash := t.hashString(val)
	t.hash.Reset()
	binary.LittleEndian.PutUint64(t.buf[:], keyHash)
	t.hash.Write(t.buf[:])
	binary.LittleEndian.PutUint64(t.buf[:], valHash)
	t.hash.Write(t.buf[:])
	return t.hash.Sum64()
}

func (t *tag) hashString(key string) uint64 {
	t.hash.Reset()
	//t.hash.Write(*(*[]byte)(unsafe.Pointer(&key)))
	t.hash.Write([]byte(key))
	return t.hash.Sum64()
}

func (t *tag) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	t.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

func (t *tag) MarshalEasyJSON(w *jwriter.Writer) {
	if len(t.values) == 0 {
		w.RawString("null")
	} else {
		w.RawString("{")
		childFirst := true
		for k, v := range t.values {
			if !childFirst {
				w.RawByte(',')
			}
			childFirst = false
			w.String(k)
			w.RawByte(':')
			w.String(v)
		}
		w.RawString("}")
	}
}
