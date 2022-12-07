package tykky

import "fmt"

var colorReset string = "\033[0m"
var colorRed string = "\033[31m"
var colorGreen string = "\033[32m"
var colorYellow string = "\033[33m"
var colorBlue string = "\033[34m"
var colorPurple string = "\033[35m"
var colorCyan string = "\033[36m"
var colorWhite string = "\033[37m"

func PrintInfo(msg string) {
	fmt.Printf("[ %sINFO%s ] %s\n", colorBlue, colorReset, msg)

}
