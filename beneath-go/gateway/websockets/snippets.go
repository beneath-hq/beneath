package websockets

// very helpful: https://github.com/eranyanay/1m-go-websockets
// scaling one server:
// problem 1. go kernel file limit is 1024 open files at once; each socket is represented by a file.
// solution 1: override the artificial file limit with a much larger number
// problem 2. at relatively small numbers of connections it will use up tons of RAM. in the example, 50k connections used up 1 GB of RAM.; each connection consumes ~20KB; 1M connections would take 20 GB of RAM
// solution 2: use Epoll (async I/O) to reduce the # of goroutines; reuse goroutines and reduce memory footprint; use epolls to do this; in the example, goroutines reduced from 50k to 5; now using 600 MB of RAM
// problem 3: buffers allocations; don't need a bufio connection because we are doing all input/output asynchronously; gorilla/websocket keeps a reference to the underlying buffers given by Hijack()
// solution 3: use more performant websockets library: github.com/gobwas/ws; this reduces buffer allocations; in the example, now using 60 MB of RAM for 50k connections.
// problem 4: conntrack table is full
// solution 4: change the server's settings to increase the cap of the total concurrent connections in the operating system

/*
// solution 1 to override resource limitations:
// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
*/
