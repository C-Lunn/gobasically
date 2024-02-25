package main

import (
	"basicallygo/context"
	"bufio"
	"os"
	"os/signal"
	"strings"
)

type t struct {
}

func (t t) Printline(s string) {
	println(s)
}

func main() {
	// 	prog :=
	// 		`10 GOTO 60
	// 20 PRINT "HELLO"
	// 30 GOTO 20
	// 40 PRINT "WORLD"
	// 50 GOTO 40
	// 60 GOTO 50`

	term := t{}

	cont := context.New(term)

	// Get user input, add it to the input buffer, then run when user types RUN
	// The input buffer is a string containing the whole program, but read one user input at a time
	// This is a simple way to simulate a user typing in a program

	interrupt := make(chan os.Signal, 1)
	quit := make(chan bool)

	signal.Notify(interrupt, os.Interrupt)

	go repl(interrupt, quit, cont)

	<-quit
}

func repl(interrupt chan os.Signal, quit chan bool, cont *context.Context) {
	prog :=
		`10 LET A = 1
20 LET B = 2
30 LET C = A + B
35 C = C + 1
36 IF C % 2 == 0 THEN PRINT C, "IS EVEN" ELSE PRINT C, "IS ODD" END
50 GOTO 35`
	// split prog
	lines := strings.Split(prog, "\n")
	for _, line := range lines {
		cont.Accept_line(line)
	}
	// cont.Accept_line("20 GOTO 10")
	done := make(chan bool)
	go cont.Run(interrupt, done)
	<-done
	reader := bufio.NewReader(os.Stdin)
	println("READY")
	for {
		// Read user input
		inp, _ := reader.ReadString('\n')
		// ignore empty
		if inp == "\n" {
			continue
		}
		//remove newline
		inp = inp[:len(inp)-1]
		if strings.ToUpper(inp) == "RUN" {
			// Run the program
			done := make(chan bool)
			go cont.Run(interrupt, done)
			<-done
		} else if strings.ToUpper(inp) == "LIST" {
			// Print the input buffer
			cont.List()
		} else if strings.ToUpper(inp) == "QUIT" {
			quit <- true
			break
		} else {
			cont.Accept_line(inp)
		}
	}
}
