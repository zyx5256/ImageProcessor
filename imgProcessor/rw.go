package imgProcessor

import (
	"image"
	"os"
	"image/png"
)

type Task struct {
	inPath string
	outPath string
	effects []string
	img *PNGImage
}

func NewTask(inPath string, outPath string, effects []string, img *PNGImage) *Task {
	return &Task{inPath : inPath, outPath : outPath, effects : effects, img : img}
}

// Load loads input image for task
func (task *Task) load() {

	inReader, err := os.Open(task.inPath)

	if err != nil {
		panic(err)
	}
	defer inReader.Close()

	inImg, err := png.Decode(inReader)

	if err != nil {
		panic(err)
	}

	inBounds := inImg.Bounds()

	outImg := image.NewRGBA64(inBounds)

	task.img = &PNGImage{in : inImg, out : outImg}
}

// sequentially load image for each task in tasks
func SeqLoad(tasks []Task) {
	
	for i := 0; i < len(tasks); i++ {

		tasks[i].load()
	}
}

// sequentially save image for each task in tasks
func SeqWrite(tasks []Task) {
	
	for i := 0; i < len(tasks); i++ {

		tasks[i].img.Save(tasks[i].outPath)
	}
}

func SeqProcess(tasks []Task)  {

	for i := 0; i < len(tasks); i++ {

		task := tasks[i]
		for _ , effect := range task.effects {

			task.img.Apply(effect, task.img.in.Bounds())
			task.img.in = task.img.out
		}
	}
}

// ParaLoad loads tasks from input stream to task stream
func ParaLoad(inputStream <-chan Task, taskStream chan<- Task) {
			
	for input := range inputStream {
		
		input.load()
		taskStream <- input
	}
}

// ParaWrite writes tasks from output stream and put it into results stream
func ParaWrite(outputStream <-chan Task, results chan<- Task) {

	for output := range outputStream {

		output.img.Save(output.outPath)
		results <- output
	}
}

// wait makes the main thread wait until all tasks are done
func Wait(results <-chan Task, numOfResult int) {

	for i := 0; i < numOfResult; i++ {
		<- results
	}
}

// stageBarrier blocks pipeline thread until every worker is done with their work
func stageBarrier(numOfWorker int, stageOut <-chan Task) {

	for i := 0; i < numOfWorker; i++ {
		<- stageOut
	}
}

func Pipeline(taskStream <-chan Task, outputStream chan<- Task, numOfWorker int) {

	for task := range taskStream {
		//fmt.Println("processing task:\n" + task.inPath)
		for _, effect := range task.effects {

			stageOut := make(chan Task, 2 * numOfWorker) // double the buffer size to ensure safty

			bounds := task.img.in.Bounds()
			len := bounds.Max.Y / numOfWorker
			
			for i:= 0; i < numOfWorker; i++ {

				// split image boundary into numOfWorker parts to run
				grid := image.Rectangle{Min : image.Point{0, i * len}, Max : image.Point{bounds.Max.X, (i + 1) * len}}
				if i == numOfWorker - 1 {
					grid.Max.Y = bounds.Max.Y
				}

				// spwan worker
				go func() {
					task.img.Apply(effect, grid)
					stageOut <- task
				}()
			}
			// wait until all workers are done
			stageBarrier(numOfWorker, stageOut)
			close(stageOut)

			// update task
			task.img.in = task.img.out
		}
		outputStream <- task
	}
}