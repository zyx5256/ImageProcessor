# Implementation Details
    1. imgProcessor.go:
      This file is like a physical layer. It implements functions like loading an image from local file, performing specific image effect, etc.

    2. rw.go:
      This file is like a logic layer. It implements two ways of processing certain amount of  Tasks: sequential and parallel. An Task is to perform series of effects on a ceratin image.
      
        2.1 Sequential processing: 
          2.1.1 SeqLoad(tasks []Task): loads the image for each Task;
          2.1.2 SeqProcess: sequentially perform each effect for the image as each Task specified;
          2.1.3 SeqWrite: writes output image for each task.

        2.2 Parallel processing:
          2.2.1 ParaLoad(inputStream <-chan Task, taskStream chan<- Task): 
            fetches available Task from inputStream and load the image of that Task, and then put the Task into taskStream.
          2.2.2 Pipeline(taskStream <-chan Task, outputStream chan<- Task, numOfWorker int): 
            fetches available Task from taskStream and perform image effects. At each effect stage, it spwan N threads to perform the effct. After all effects are applied, the Task will be put into outputStream. At each effect stage, there's a stage barrier that ensures each of N threads has done their work before entering the next stage.
          2.2.3 ParaWrite(outputStream <-chan Task, results chan<- Task): 
            fetches available Task from outputStream and wirte the image  of that Task to local directory, and then put the Task into results.
          2.2.4 Wait(results <-chan Task, numOfResult int): 
            this is another barrier that ensures the total number of Task processed is equal to numOfResult, which is the number of command in the command file.

    3. editor.go:
      This file is like a view layer. It handles user inputs and parses those inputs into a slice of Tasks.

# Hardware Overview
    Model Name:	iMac
    Model Identifier:	iMac15,1
    Processor Name:	Intel Core i7
    Processor Speed:	4 GHz
    Number of Processors:	1
    Total Number of Cores:	4
    L2 Cache (per Core):	256 KB
    L3 Cache:	8 MB
    Memory:	16 GB
    Boot ROM Version:	IM151.0217.B00
    SMC Version (system):	2.23f11
    Serial Number (system):	C02Q418VFY14
    Hardware UUID:	D507A51F-7C8D-5951-BACE-59709F47947F

# Result Graph

![Number of Threads vs Speedup](https://mit.cs.uchicago.edu/mpcs52060-aut-19/zyx/raw/master/src/mpcs52060/proj2/speedup.jpeg)

# SAQs
    1. What are the hotspots and bottlenecks in your sequential program? Were you able to parallelize the hotspots and/or remove the bottlenecks in the parallel version?
      There is 1 hotspot: performing image effects.
      There are 2 bottlenecks: loading and writing image file.
      I parallize the hotspot by spwan N threads to perform each image effect(data decomposition);
      I remove the bottlenecks by spwan N/5 threads to parallel reading, prcessing and writing images(functional decomposition).
    2 Describe the granularity in your implementation. Are you using a coarse-grain or fine-grain granularity? Explain.
      Coarse-grain granularity. Relatively large amounts of work are done between communication events
    3. Does the image size being processed have any effect on the performance? For example, processing csv files with the same number of images but one file has very large image sizes, whereas the other has smaller image sizes.
      Yes there is. If the large image is processed at the beginning, then the finish time for images that come after it will increase considerably.


