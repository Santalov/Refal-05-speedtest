package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	total             float32
	totalRefal        float32
	builtin           float32
	linearResult      float32
	linearPattern     float32
	openELoop         float32
	TandEVarCopy      float32
	repeatedEvarMatch float32
	stepCount         int
	memoryUsedNodes   int
}

func getProfilerDurationVal(s string, alias string) float32 {
	accuracy := 3
	i := strings.Index(s, alias)
	if i == -1 {
		return float32(0)
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
	return float32(value)
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

func calcAverage(results []*profilerOutput) *profilerOutput {
	res := profilerOutput{}
	for _, p := range results {
		res.total += p.total
		res.totalRefal += p.totalRefal
		res.builtin += p.builtin
		res.linearResult += p.linearResult
		res.linearPattern += p.linearPattern
		res.openELoop += p.openELoop
		res.TandEVarCopy += p.TandEVarCopy
		res.repeatedEvarMatch += p.repeatedEvarMatch
		res.stepCount += p.stepCount
		res.memoryUsedNodes += p.memoryUsedNodes
	}
	n := float32(len(results))
	res.total /= n
	res.totalRefal /= n
	res.builtin /= n
	res.linearResult /= n
	res.linearPattern /= n
	res.openELoop /= n
	res.TandEVarCopy /= n
	res.repeatedEvarMatch /= n
	res.stepCount /= int(n)
	res.memoryUsedNodes /= int(n)
	return &res
}

func pow2(v float32) float32 {
	return v * v
}

func calcDifferenceQuads(results []*profilerOutput, avg *profilerOutput) []*profilerOutput {
	answer := make([]*profilerOutput, len(results))
	for i, p := range results {
		answer[i] = new(profilerOutput)
		answer[i].total = pow2(avg.total - p.total)
		answer[i].totalRefal = pow2(avg.totalRefal - p.totalRefal)
		answer[i].builtin = pow2(avg.builtin - p.builtin)
		answer[i].linearResult = pow2(avg.linearResult - p.linearResult)
		answer[i].linearPattern = pow2(avg.linearPattern - p.linearPattern)
		answer[i].openELoop = pow2(avg.openELoop - p.openELoop)
		answer[i].TandEVarCopy = pow2(avg.TandEVarCopy - p.TandEVarCopy)
		answer[i].repeatedEvarMatch = pow2(avg.repeatedEvarMatch - p.repeatedEvarMatch)
		answer[i].stepCount = 0
		answer[i].memoryUsedNodes = 0
	}
	return answer
}

func root2(v float32) float32 {
	return float32(math.Sqrt(float64(v)))
}

func calcRoots(p *profilerOutput) *profilerOutput {
	res := profilerOutput{}
	res.total = root2(p.total)
	res.totalRefal = root2(p.totalRefal)
	res.builtin = root2(p.builtin)
	res.linearResult = root2(p.linearResult)
	res.linearPattern = root2(p.linearPattern)
	res.openELoop = root2(p.openELoop)
	res.TandEVarCopy = root2(p.TandEVarCopy)
	res.repeatedEvarMatch = root2(p.repeatedEvarMatch)
	res.stepCount = p.stepCount
	res.memoryUsedNodes = p.memoryUsedNodes
	return &res
}

func main() {
	if !processCommandLineArgs() {
		return
	}
	if numberOfTests <= 0 {
		fmt.Println("numberOfTest should be greater than 0")
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
