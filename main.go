package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var pathToExecutable string
var pathToInputFile string
var numberOfTests = 21

func processCommandLineArgs() (ok bool) {
	if len(os.Args) == 1 || len(os.Args) > 4 {
		fmt.Println("pass pathToExecutable [pathToInputFile] [numberOfTests]" +
			" as command line arguments")
		return false
	}
	pathToExecutable = os.Args[1]
	if len(os.Args) >= 3 {
		numberOfTests64, err := strconv.ParseInt(os.Args[2], 10, 32)
		if err != nil {
			pathToInputFile = os.Args[2]
			if len(os.Args) == 4 {
				numberOfTests64, err := strconv.ParseInt(os.Args[3], 10, 32)
				if err != nil {
					fmt.Println("the second argument should either be" +
						" a path to input file or number of tests")
					return false
				}
				numberOfTests = int(numberOfTests64)
			}
		} else {
			numberOfTests = int(numberOfTests64)
			pathToInputFile = os.Args[3]
		}
	}
	return true
}

type profilerOutput struct {
	total             time.Duration
	totalRefal        time.Duration
	builtin           time.Duration
	linearResult      time.Duration
	linearPattern     time.Duration
	openELoop         time.Duration
	TandEVarCopy      time.Duration
	repeatedEvarMatch time.Duration
	stepCount         int
	memoryUsedNodes   int
}

func getProfilerDurationVal(s string, alias string) time.Duration {
	accuracy := 3
	i := strings.Index(s, alias)
	if i == -1 {
		return time.Duration(0)
	}
	j := i + len(alias)
	for ; s[j] == ' '; j++ {
	}
	k := j
	for ; s[k] != '.'; k++ {
	}
	valueStr := s[j : k+accuracy]
	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		panic(err)
	}
	return time.Duration(value * math.Pow10(9))
}

func getProfilerCountVal(s string, alias string) int {
	i := strings.Index(s, alias)
	if i == -1 {
		return 0
	}
	j := i + len(alias)
	for ; s[j] == ' '; j++ {
	}
	k := j
	for ; s[k]-'0' < 10; k++ {
	}
	valueStr := s[j:k]
	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(value)
}

func parseProfilerOutput(s string) *profilerOutput {
	res := profilerOutput{}
	res.total = getProfilerDurationVal(s, "Total program time:")
	res.totalRefal = getProfilerDurationVal(s, "(Total refal time):")
	res.builtin = getProfilerDurationVal(s, "Builtin time:")
	res.linearResult = getProfilerDurationVal(s, "Linear result time:")
	res.linearPattern = getProfilerDurationVal(s, "Linear pattern time:")
	res.openELoop = getProfilerDurationVal(s, "Open e-loop time (clear):")
	res.TandEVarCopy = getProfilerDurationVal(s, "t- and e-var copy time:")
	res.repeatedEvarMatch = getProfilerDurationVal(s, "Repeated e-var match time (inside e-loops):")
	res.stepCount = getProfilerCountVal(s, "Step count")
	res.memoryUsedNodes = getProfilerCountVal(s, "Memory used")
	return &res
}

func main() {
	if !processCommandLineArgs() {
		return
	}
	results := make([]*profilerOutput, numberOfTests)
	for i := 0; i < numberOfTests; i++ {
		cmd := exec.Command(pathToExecutable)
		if pathToExecutable != "" {
			file, err := os.Open(pathToInputFile)
			if err != nil {
				fmt.Println("cannot open input file", pathToInputFile, err)
				return
			}
			cmd.Stdin = file
		}
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("executable have returned a error", err)
		}
		results[i] = parseProfilerOutput(string(stdoutStderr))
		fmt.Println(*results[i])
	}
}
