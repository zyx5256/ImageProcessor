package main

import (
	"os"
	"../imgProcessor"
	"bufio"
	"encoding/json"
	"reflect"
	"runtime"
	"math"
	"flag"
	"fmt"
)

// Parser parses command file to a slice of *task
func Parser() []imgProcessor.Task {

	var tasks []imgProcessor.Task

	readBuffer := bufio.NewReader(os.Stdin)

	// read line while has next line
	var row string
	var err error
	for err == nil {

		// read the next line
		row, err = readBuffer.ReadString('\n')
		if len(row) == 0 {
			break
		}

		var input interface{}
		// parse json to Task
		b := []byte(row)
		json.Unmarshal(b, &input)
		m := input.(map[string]interface{})

		// parse effects
		effects := make([]string, 0)
		rEffect := reflect.ValueOf(m["effects"])
		for i := 0; i < rEffect.Len(); i++ {
			effects = append(effects, rEffect.Index(i).Interface().(string))
		}
		curDir, _ := os.Getwd()
		task := imgProcessor.NewTask(curDir + "/" + m["inPath"].(string), curDir + "/" + m["outPath"].(string), effects, nil)

		// append to tasks
		tasks = append(tasks, *task)
	}
	return tasks
}

func sequentialProcess(tasks []imgProcessor.Task) {

	// load image to each task
	imgProcessor.SeqLoad(tasks)

	// process each task with effect blur
	imgProcessor.SeqProcess(tasks)

	//Saves the image to a new file
	imgProcessor.SeqWrite(tasks)
}

func parallelProcess(tasks []imgProcessor.Task, numOfThread int) {

	// create Ceil(numOfThread/5) loader to load images
	inputStream := make(chan imgProcessor.Task)
	taskStream := make(chan imgProcessor.Task)
	outputStream := make(chan imgProcessor.Task)
	results := make(chan imgProcessor.Task, 2*len(tasks)) // double the buffer size to ensure safty

	defer close(inputStream)
	defer close(taskStream)
	defer close(outputStream)

	// Ceil(numOfThread / 5) threads to read, process and write, respectively
	for i := 0; i < int(math.Ceil(float64(numOfThread) / 5)); i++ {

		go imgProcessor.ParaLoad(inputStream, taskStream)
		go imgProcessor.Pipeline(taskStream, outputStream, numOfThread)
		go imgProcessor.ParaWrite(outputStream, results)
	}

	// put tasks to input stream
	for _, task := range tasks {

		inputStream <- task
	}
	imgProcessor.Wait(results, len(tasks))
}

func main() {

	//Read tasks from to command file
	numOfThreads := flag.Int("p", -1, "number of threads")
	flag.Parse()
	
	if *numOfThreads < -1 {

		fmt.Println("Usage: editor [-p=[number of threads]]")
		fmt.Println("	-p=[number of threads] = An optional flag to run the editor in its parallel version.Call and pass the runtime.GOMAXPROCS(...) function the integer specified by [number of threads].")
		return
	}

	tasks := Parser()
	if *numOfThreads == -1 {

			fmt.Println("Begin sequential process...")
			sequentialProcess(tasks)

	} else {

		fmt.Printf("Begin parallel process, maximum numOfThreads = %v ...\n", *numOfThreads)
		runtime.GOMAXPROCS(*numOfThreads)
		parallelProcess(tasks, *numOfThreads)

	} 
}
