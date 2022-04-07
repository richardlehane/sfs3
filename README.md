# sfs3

This package is a demonstration of how to do range reading of an S3 object with [siegfried](https://github.com/richardlehane/siegfried).

This isn't an optimised solution and the approach could be improved, e.g.:

  - with a larger internal buffer you could avoid doing range requests for every call of the Slice method
  - you could add logging to determine how many requests are made, how many bytes are downloaded
  - you could add throttling, time outs, and maximum bounds on number/size of requests.
