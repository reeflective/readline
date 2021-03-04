package main

import (
	"fmt"
	"os"

<<<<<<< HEAD
	"github.com/bishopfox/sliver/client/readline"
=======
	"github.com/maxlandon/readline"
>>>>>>> 611c6fb333d138b32958059c075a2d21c7ca09ae
)

func main() {
	readline.MakeRaw(int(os.Stdin.Fd()))

	for {
		b := make([]byte, 1024)
		i, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}

		fmt.Println(b[:i])
	}
}
