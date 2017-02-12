package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/stxmendez/strfile"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var printHeader, random bool
	flag.BoolVar(&printHeader, "printHeader", false, "show header")
	flag.BoolVar(&random, "random", false, "show a random message")

	flag.Parse()

	n := flag.NArg()
	if n != 2 {
		log.Fatal("Need to specify the data file and the strings file")
	}

	strPath := flag.Arg(0)
	datPath := flag.Arg(1)
	reader, err := strfile.NewStrFileReader(strPath, datPath)
	if err != nil {
		log.Fatal(err)
	}

	h, err := reader.Header()
	if err != nil {
		log.Fatal(err)
	}

	if printHeader {
		fmt.Printf("Header: %#v\n", *h)
	}

	if random {
		str, err := reader.String(rand.Intn(int(h.Numstr)))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(str)
	}
}
