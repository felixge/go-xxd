# go-xxd

This repository contains my question to [How can I improve the performance of
my xxd
port?](http://www.reddit.com/r/golang/comments/2s1zn1/how_can_i_improve_the_performance_of_my_xxd_port/)
on reddit.

The result is a Go version of xxd that outperforms the native versions on OSX
10.10.1 / Ubuntu 14.04 (inside VirtualBox), see benchmarks below. However, that
is not impressive, given that none of the usual xxd flags are supported.

What is interesting however, are the steps to get there:

* Make the code testable and compare against output of native xxd using test/quick: https://github.com/felixge/go-xxd/commit/90262b3dcdc518ca3eaec7171aa14d74d95f34b8
* Fix bugs: https://github.com/felixge/go-xxd/commit/e9ebeb0abdf78f6e7729fdbfc68842b3a86ee0a3, https://github.com/felixge/go-xxd/commit/120804574f12033999f23e6cf6a3b75961f14da1, https://github.com/felixge/go-xxd/commit/dab678ecf5dcb3eff345db8ac68ae6d7438f9d0e
* Buffer output to stdout: https://github.com/felixge/go-xxd/commit/69b5fe0cc7da80d374413d72892507d5e5ecaabc
* Implement a benchmark: https://github.com/felixge/go-xxd/commit/0bce954073ce92b72ed3fbcf36603c6e23852feb
* Use the benchmark + [go pprof](http://blog.golang.org/profiling-go-programs) to get ideas for optimization
* Remove unneeded printf calls: https://github.com/felixge/go-xxd/commit/473330acd320e5318e896d6408fb3d64a5b8e10b
* Optimize hex encoding and avoid allocations: https://github.com/felixge/go-xxd/commit/dce3bca200ae499e1cf57994c7592a42f66694d5, https://github.com/felixge/go-xxd/commit/a48d892de3fb625bdb3e8e367337578c766a42f5, https://github.com/felixge/go-xxd/commit/d653d2f4eeb5fa41e907f260bd95c205bd8e1ff7, https://github.com/felixge/go-xxd/commit/0d3ae7f0be863a138fab3fd2dd89208073b61c7f
* Avoid type casts: https://github.com/felixge/go-xxd/commit/859ebc4489ce81edc0681881700b0f22754943f0

You can also follow along by looking at the commit history: https://github.com/felixge/go-xxd/commits/master

## OSX 10.10.1:

### xxd native:

```
$ time xxd image.jpg > /dev/null

real	0m0.205s
user	0m0.202s
sys	0m0.003s
```

### xxd.go (original version from reddit):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m5.914s
user	0m3.598s
sys	0m2.318s
```

### xxd.go (optimized):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m0.138s
user	0m0.133s
sys	0m0.004s
```

## Ubuntu 14.04 (inside VirtualBox):

### xxd native:

```
$ time xxd image.jpg > /dev/null

real	0m0.273s
user	0m0.017s
sys	0m0.231s
```

### xxd.go (original version from reddit):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m5.856s
user	0m3.517s
sys	0m1.897s
```

### xxd.go (optimized):

```
$ go build xxd.go && time ./xxd image.jpg > /dev/null

real	0m0.233s
user	0m0.021s
sys	0m0.207s
```
