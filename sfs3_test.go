package sfs3

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
)

func Example() {
	// make a new siegfried
	sf, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatal(err)
	}
	// set up AWS session/ service
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	svc := s3.New(sess)
	// make a new Object
	obj, err := New(svc, os.Getenv("AWS_BUCKET"), os.Getenv("AWS_GIF"))
	if err != nil {
		log.Fatal(err)
	}
	// pass the Object to sf to get a siegfried buffer
	buf, err := sf.Buffer(obj)
	// use the IdentifyBuffer method to do the identification
	ids, err := sf.IdentifyBuffer(buf, err, os.Getenv("AWS_GIF"), obj.MIME)
	if err != nil {
		log.Fatal(err)
	}
	// the Object keeps count of the number of fetches and bytes transferred
	log.Printf("Performed %d fetches and retrieved %d bytes. The file size is %d bytes.", obj.RequestCount, obj.ByteCount, obj.Sz)
	// print out the ids
	for _, id := range ids {
		// sf Label decorates our id fields with labels
		for _, kv := range sf.Label(id) {
			fmt.Printf("%s: %s\n", kv[0], kv[1])
		}
	}
	// Output:
	// namespace: pronom
	// id: fmt/4
	// format: Graphics Interchange Format
	// version: 89a
	// mime: image/gif
	// class: Image (Raster)
	// basis: extension match gif; mime match image/gif; byte match at [[0 6] [1001717 1]]
	// warning:
}

func TestIDs(t *testing.T) {
	sf, err := siegfried.Load(config.Signature())
	if err != nil {
		log.Fatal(err)
	}
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	svc := s3.New(sess)
	expect := []string{"fmt/4", "fmt/43", "fmt/199", "fmt/11"}
	for aws_idx, aws_object := range []string{"AWS_GIF", "AWS_JPG", "AWS_MP4", "AWS_PNG"} {
		obj, err := New(svc, os.Getenv("AWS_BUCKET"), os.Getenv(aws_object))
		if err != nil {
			t.Fatal(err)
		}
		buf, err := sf.Buffer(obj)
		ids, err := sf.IdentifyBuffer(buf, err, os.Getenv(aws_object), obj.MIME)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("%s: performed %d fetches and retrieved %d bytes. The file size is %d bytes.", aws_object, obj.RequestCount, obj.ByteCount, obj.Sz)
		// check the id
		if ids[0].String() != expect[aws_idx] {
			t.Fatalf("Expected %s, got %s", expect[aws_idx], ids[0].String())
		}
	}
}
