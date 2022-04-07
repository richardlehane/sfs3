// Faster identification of AWS S3 files with siegfried
//
// For background, see https://github.com/richardlehane/siegfried/issues/169
//
// tldr: siegfried is slow at identifying objects in AWS S3
// This is because it defaults to a stream reader, which needs a full file read to reach the end-of-file (many PRONOM signatures have EOF sequences)
//
// By implementing siegfried's internal "source" interface (https://github.com/richardlehane/siegfried/blob/main/internal/siegreader/external.go) for AWS S3,
// we can enable sf to more selectively scan the file. This reduces the need for full file scans and downloads.
//
// For reference, the source interface is:
//
// type source interface {
//	 IsSlicer() bool
//	 Slice(off int64, l int) ([]byte, error)
//	 EofSlice(off int64, l int) ([]byte, error)
//	 Size() int64
// }
//
// Example usage of this package:
//
// func IdentifyS3(sf *siegfried.Siegfried, svc *s3.S3, bucket string, key string) ([]core.Identification, error) {
//	 obj, err := sfs3.New(svc, bucket, key)
//	 if err != nil {
//		 return nil, err
//	 }
//	 buf, err := sf.Buffer(obj)
//	 if err != nil {
//		return nil, err
//	 }
//	 return sf.IdentifyBuffer(buf, err, key, obj.MIME)
// }

package sfs3

import (
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)
var counter int
// Object uses range requests to incrementally read an S3 object
type Object struct {
	Svc          *s3.S3 // AWS Service Client
	Sz           int64
	MIME         string
	RequestInput *s3.GetObjectInput

	buf []byte
	off int64
	l   int
}

// New creates a new Object. It makes one HeadObject request to fill the Size and MIME fields.
// If you already know the size and MIME Type, make the object yourself!
func New(svc *s3.S3, bucket string, key string) (*Object, error) {
	// first get the size and MIME type of the object
	head, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &Object{
		Svc:  svc,
		Sz:   *head.ContentLength,
		MIME: *head.ContentType,
		RequestInput: &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	}, nil
}

// IsSlicer declares that Object implements the source interface
func (o *Object) IsSlicer() bool {
	return true
}

// Slice returns a byte slice at offset off, with length l
func (o *Object) Slice(off int64, l int) ([]byte, error) {
	counter++
	log.Printf("Slice called, count: %d, off: %d, len: %d\n", counter, off, l)
	// if we already have the bytes in the buf slice, return immediately
	if off >= o.off && off+int64(l) <= o.off+int64(o.l) {
		start := int(off - o.off)
		return o.buf[start : start+l], nil
	}
	if off >= o.Sz {
		return nil, io.EOF
	}
	// calculate the range
	var err error
	if off+int64(l) > o.Sz {
		err = io.EOF
		l = int(o.Sz - off)
	}
	o.RequestInput.Range = aws.String(fmt.Sprintf("bytes=%d-%d", off, off+int64(l)))
	// now GetObject
	out, e := o.Svc.GetObject(o.RequestInput)
	log.Println("Performing a range request")
	if e != nil {
		return nil, e
	}
	// resize the buf if necessary
	if len(o.buf) < l {
		o.buf = make([]byte, l)
	}
	n, e := out.Body.Read(o.buf)
	if n < 1 {
		return nil, e
	}
	o.off = off
	o.l = n
	return o.buf[:n], err
}

// EofSlice returns a slice from the end of the file at offset off, with length l
func (o *Object) EofSlice(off int64, l int) ([]byte, error) {
	if off >= o.Sz {
		return nil, io.EOF
	}
	var err error
	if l > int(o.Sz-off) {
		l, off, err = int(o.Sz-off), 0, io.EOF
	} else {
		off = o.Sz - off - int64(l)
	}
	slc, err1 := o.Slice(off, l)
	if err1 != nil {
		err = err1
	}
	return slc, err
}

// Size returns the Object's content length
func (o *Object) Size() int64 {
	return o.Sz
}

// Read ensures we are an io.Reader as well. This method should never be used within siegfried
func (o *Object) Read(b []byte) (int, error) {
	var off int64
	// if not the first read, increment the offset
	if o.l > 0 {
		off = o.off + int64(o.l)
	}
	// now get a slice
	slc, err := o.Slice(off, len(b))
	if slc == nil {
		return 0, err
	}
	return copy(b, slc), err
}
