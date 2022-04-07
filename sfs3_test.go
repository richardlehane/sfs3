package sfs3

import (
  "testing"
  "log"
  "fmt"
  
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
  "github.com/richardlehane/siegfried"
)

func ExampleSfs3() {
  // make a new siegfried
  sf, err := siegfried.Load("latest") // available at https://www.itforarchivists.com/siegfried/latest
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
  obj, err := New(svc, os.Getenv("AWS_BUCKET"), os.Getenv("AWS_OBJECT"))
  if err != nil {
    log.Fatal(err)
  }
  // pass the Object to sf to get a siegfried buffer
  buf, err := sf.Buffer(obj)
  // use the IdentifyBuffer method to do the identification
  ids, err := sf.IdentifyBuffer(buf, err, os.Getenv("AWS_OBJECT"), obj.MIME)
  if err != nil {
    log.Fatal(err)
  }
  // print out the ids
  for _, id := range ids {
    // sf Label decorates our id fields with labels
    for _, kv := range sf.Label(id) {
      fmt.Printf("%s: %s\n", kv[0], kv[1])
    }
  }
}
