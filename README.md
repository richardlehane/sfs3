# sfs3

This package is a demonstration of how to do range reading of an S3 object with [siegfried](https://github.com/richardlehane/siegfried).

It works by implementing siegfried's [source](https://github.com/richardlehane/siegfried/blob/main/internal/siegreader/external.go) interface.

Still a work in progress: this isn't an optimised solution, is currently slower and more costly than full reads for most file types, and isn't recommend for use yet.

# Usage

```go
func Example() {
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
  obj, err := sfs3.New(svc, os.Getenv("AWS_BUCKET"), os.Getenv("AWS_OBJECT"))
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
  // Output:
  // namespace: pronom
  // id: fmt/43
  // format: JPEG File Interchange Format
  // version: 1.01
  // mime: image/jpeg
  // basis: extension match jpg; mime match image/jpeg; byte match at [[0 14] [98409 2]]
  // warning:
}
```

