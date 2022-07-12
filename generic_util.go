package tykky

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func WriteShellScript(filename string, content []string) {
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		panic("Failed to create shell script")
	}
	f.WriteString("#!/bin/bash\n")
	f.WriteString("set -e\n")
	f.WriteString("set -u\n")
	for _, s := range content {
		f.WriteString(s + "\n")
	}
	os.Chmod(filename, 0700)
}

func ResolvePath(p string) string {
	if p[0] == '/' {
		return path.Clean(p)
	} else {
		currentDir, _ := os.Getwd()
		return path.Clean(currentDir + "/" + p)
	}

}

type outputWriter struct {
	streamIn  io.Reader
	streamOut io.Writer
	// How many lines to display
	// -1 normal print
	maxNumberOfLines int
	fullOutPutFile   string
}

var PlainOutputWriterNoLog = outputWriter{nil, os.Stdout, -1, ""}
var ScrollingOutputWriterNoLog = outputWriter{nil, os.Stdout, 20, ""}

func (o outputWriter) outputCanBeRolled() bool {
	return o.maxNumberOfLines > 0 && !OutputIsRedirected()
}

func EraseLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("\033[2K\033[1A\033[2K")
	}
}

func min(a int, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func constructCommand(fullCommand string) *exec.Cmd {
	arguments := strings.Split(fullCommand, " ")
	commandName := arguments[0]
	cmd := exec.Command(commandName, arguments[1:]...)
	return cmd
}

func RunSilentCommand(fullCommand string) {
	nothing := func(*exec.Cmd, *outputWriter) {}
	_Run(fullCommand, outputWriter{}, nothing, nothing)
}

func RunPlainOutput(fullCommand string) {
	before := func(e *exec.Cmd, o *outputWriter) {
		e.Stdout = os.Stdout
		e.Stderr = os.Stderr
	}
	after := func(e *exec.Cmd, o *outputWriter) {
	}
	_Run(fullCommand, outputWriter{}, before, after)
}

func _Run(fullCommand string, commandOutput outputWriter, fBefore func(*exec.Cmd, *outputWriter), fAfter func(*exec.Cmd, *outputWriter)) {
	cmd := constructCommand(fullCommand)
	cmd.Stdin = os.Stdin
	fBefore(cmd, &commandOutput)
	cmdErr := cmd.Start()
	if cmdErr != nil {
		fmt.Printf(cmdErr.Error())
	}
	fAfter(cmd, &commandOutput)
	cmd.Wait()
}
func RunPseudoInteractiveCommand(fullCommand string, commandOutput outputWriter, stdinContent string) {
	var b bytes.Buffer
	b.Write([]byte(stdinContent))
	before := func(e *exec.Cmd, o *outputWriter) {
		cmdReader, _ := e.StdoutPipe()
		e.Stderr = e.Stdout
		o.streamIn = cmdReader
		e.Stdin = &b

	}
	after := func(e *exec.Cmd, o *outputWriter) {
		o.RawProgramPrint()
	}
	//cmd.Stdin = os.Stdin
	_Run(fullCommand, commandOutput, before, after)
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [yes|no]:\n", s)
		response, _ := reader.ReadString('\n')

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "yes" {
			return true
		} else if response == "no" {
			return false
		}
		fmt.Printf("Type either yes or no\n")
	}
}

func RunCommand(fullCommand string, commandOutput outputWriter) {
	before := func(e *exec.Cmd, o *outputWriter) {
		//stdout, _ := e.StdoutPipe()
		//o.stream = stdout
		//stderr, _ := e.StderrPipe()
		//o.stream = io.MultiReader(stdout, stderr)
		cmdReader, _ := e.StdoutPipe()
		e.Stderr = e.Stdout
		o.streamIn = cmdReader
	}
	after := func(e *exec.Cmd, o *outputWriter) {
		if o.outputCanBeRolled() {
			o.FollowProgramOutput()
		} else {
			o.RawProgramPrint()
		}
	}
	_Run(fullCommand, commandOutput, before, after)
}

func OutputIsRedirected() bool {
	o, _ := os.Stdout.Stat()
	return (o.Mode() & os.ModeCharDevice) != os.ModeCharDevice
}

func PrintProgramOutput(lines []string, nLines int, maxLines int, pl int) {
	EraseLines(pl)
	fmt.Printf("################\n")
	fmt.Printf("%s\n", strings.Join(lines[0:min(nLines, maxLines)], "\n"))
	fmt.Printf("################\n")
}

func (o outputWriter) RawProgramPrint() {
	scanner := bufio.NewScanner(o.streamIn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		msg := scanner.Text()
		io.WriteString(o.streamOut, fmt.Sprintf("%s\n", msg))
	}

}

func (o outputWriter) FollowProgramOutput() {
	scanner := bufio.NewScanner(o.streamIn)
	scanner.Split(bufio.ScanLines)
	var printLines = make([]string, o.maxNumberOfLines)
	for i := range printLines {
		printLines[i] = ""
	}
	var out *os.File
	defer out.Close()
	if o.fullOutPutFile != "" {
		out, _ = os.Create(o.fullOutPutFile)
	}
	lineCounter := 0
	numPrintedLines := 0
	t1 := time.Now()
	for scanner.Scan() {
		t2 := time.Now()
		m := scanner.Text()
		// Save the full output if we are showing just a partial view
		if o.fullOutPutFile != "" {
			_, err := out.WriteString(fmt.Sprintf("%s\n", m))
			if err != nil {
				panic(err.Error())
			}
		}
		if lineCounter >= o.maxNumberOfLines {
			for i := 0; i < o.maxNumberOfLines-1; i++ {
				printLines[i] = printLines[i+1]
			}
		}
		printLines[min(o.maxNumberOfLines-1, lineCounter)] = m
		lineCounter++
		if t2.Sub(t1).Milliseconds() > 100 {
			t1 = time.Now()
			PrintProgramOutput(printLines, lineCounter, o.maxNumberOfLines, numPrintedLines)
			numPrintedLines = min(lineCounter, o.maxNumberOfLines) + 2
		}
	}
	PrintProgramOutput(printLines, lineCounter, o.maxNumberOfLines, numPrintedLines)
}

func PrintDownloadProgress(done chan int64, path string, totalSize int64) {
	stop := false
	out, _ := os.Open(path)
	defer out.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			ft, _ := out.Stat()
			fmt.Printf("\r%d/%d", ft.Size(), totalSize)
		}
		if stop {
			ft, _ := out.Stat()
			fmt.Printf("\r%d/%d", ft.Size(), totalSize)
			fmt.Printf("\nStopped printing progress\n")
			break
		}
		time.Sleep(time.Second)

	}
}

func DownloadFile(url string, filepath string, showProgress bool) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	done := make(chan int64)
	if showProgress {
		// Get the size of the file
		headResp, err := http.Head(url)
		if err != nil {
			panic(err)
		}
		defer headResp.Body.Close()
		size, err := strconv.Atoi(headResp.Header.Get("Content-length"))
		go PrintDownloadProgress(done, filepath, int64(size))
	}
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if showProgress {
		done <- 1
	}

	return nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func CreateBuildDir(basePath string) string {
	folderName := path.Clean(basePath + "/tykky-" + RandStringBytes(7))
	os.MkdirAll(folderName, os.ModePerm)
	return folderName
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
func CopyFile(src string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	io.Copy(in, out)
	return nil
}
