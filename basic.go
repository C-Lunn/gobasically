package main

import (
	"basicallygo/context"
	"basicallygo/terminal"
	"bufio"
	"os"
	"strings"
)

func main() {
	//     prog :=
	//         `10 GOTO 60
	// 20 PRINT "HELLO"
	// 30 GOTO 20
	// 40 PRINT "WORLD"
	// 50 GOTO 40
	// 60 GOTO 50`

	term := terminal.Terminal{
		Printline: func(s string) {
			println(s)
		},
	}

	cont := context.New(&term)

	// Get user input, add it to the input buffer, then run when user types RUN
	// The input buffer is a string containing the whole program, but read one user input at a time
	// This is a simple way to simulate a user typing in a program

	interrupt := make(chan os.Signal, 1)
	interrupt_bool := make(chan bool)
	quit := make(chan bool)

	//signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			select {
			case <-interrupt:
				interrupt_bool <- true
				return
			}
		}
	}()

	go repl(interrupt_bool, quit, cont)

	<-quit
}

func repl(interrupt chan bool, quit chan bool, cont *context.Context) {
	prog := `10 REM WHATEVER
20 LET A = "HELLO"
25 LET B = ""
30 FOR I = 0 TO LEN(A) + 5
41 PRINT A[I]
50 NEXT
60 PRINT B`
	// 	prog := `10 LET A = 1
	// 20 A = A + 1
	// 30 PRINT A`
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
