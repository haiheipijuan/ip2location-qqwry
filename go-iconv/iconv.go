//
// iconv.go
//
package iconv

// #cgo CFLAGS: -I/usr/local/include
// #cgo LDFLAGS: -liconv -L/usr/local/lib
// #include <iconv.h>
// #include <errno.h>
import "C"

import (
	"io"
	"unsafe"
	"bytes"
	"syscall"
)

var EILSEQ = syscall.Errno(C.EILSEQ)
var E2BIG = syscall.Errno(C.E2BIG)

const DefaultBufSize = 4096

type Iconv struct {
	Handle C.iconv_t
}

func Open(tocode string, fromcode string) (cd Iconv, err error) {
	ret, err := C.iconv_open(C.CString(tocode), C.CString(fromcode))
	if err != nil {
		return
	}
	cd = Iconv{ret}
	return
}

func (cd Iconv) Close() error {
	_, err := C.iconv_close(cd.Handle)
	return err
}

func (cd Iconv) Conv(b []byte, outbuf []byte) (out []byte, inleft int, err error) {

	//outn, inleft, err := cd.Do(b, len(b), outbuf)	
	//if err == nil && err != E2BIG {
	//	out = outbuf[:outn]
	//	return
	//}

	w := bytes.NewBuffer(nil)
	//w.Write(out)

	inleft, err = cd.DoWrite(w, b/*[len(b)-inleft:]*/, len(b)/*inleft*/, outbuf)
	out = w.Bytes()
	return
}

func (cd Iconv) ConvString(s string) string {
	var outbuf [512]byte
	s1, _, _ := cd.Conv([]byte(s), outbuf[:])
	return string(s1)
}

func (cd Iconv) Do(inbuf []byte, in int, outbuf []byte) (out, inleft int, err error) {

	if in == 0 { return }
	
	inbytes := C.size_t(in)
	inptr := &inbuf[0]

	outbytes := C.size_t(len(outbuf))
	outptr := &outbuf[0]
	_, err = C.iconv(cd.Handle,
		(**C.char)(unsafe.Pointer(&inptr)), &inbytes,
		(**C.char)(unsafe.Pointer(&outptr)), &outbytes)

	out = len(outbuf) - int(outbytes)
	inleft = int(inbytes)
	return
}

func (cd Iconv) DoWrite(w io.Writer, inbuf []byte, in int, outbuf []byte) (inleft int, err error) {

	if in == 0 { return }

	inbytes := C.size_t(in)
	inptr := &inbuf[0]

	for inbytes > 0 {
		outbytes := C.size_t(len(outbuf))
		outptr := &outbuf[0]
		_, err = C.iconv(cd.Handle,
			(**C.char)(unsafe.Pointer(&inptr)), &inbytes,
			(**C.char)(unsafe.Pointer(&outptr)), &outbytes)
		w.Write(outbuf[:len(outbuf)-int(outbytes)])
		if err != nil && err != E2BIG {
			return int(inbytes), err
		}
	}

	return 0, nil
}
