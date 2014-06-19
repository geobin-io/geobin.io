## Cross-compiled Build

This assumes you have set up the Go environment required to compile for your destination.
* For cross compilation setup, see [this blog from Dave Cheney](http://dave.cheney.net/2013/07/09/an-introduction-to-cross-compilation-with-go-1-1)

To produce a default tar.gz targeted at linux/amd64:
```bash
> go run build.go
```

If you want to build for a different OS and Arch:
```bash
> go run build.go -os myOS -arch myArch
```

As an example, to build for 32-bit windows:
```bash
> go run build.go -os windows -arch 386
```
