package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oshankkumar/tic-tac-go/server"
)

var (
	port = flag.Int("port", 8000, "port number of tcp server")
)

func main() {
	flag.Parse()
	checkError(server.Start(*port))
}

func checkError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Fatal error : %s", err)
		os.Exit(1)
	}
}
