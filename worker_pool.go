package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	workers      map[int]*Worker
	workerCount  int64
	nextWorkerID int64
	jobQueue     chan Job
	quit         chan bool
	mu           sync.RWMutex
	wg           sync.WaitGroup
}

type Worker struct {
	id       int
	pool     *WorkerPool
	jobQueue chan Job
	quit     chan bool
}

type Job struct {
	ID   int
	Data string
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		workers:  make(map[int]*Worker),
		jobQueue: make(chan Job, 100),
		quit:     make(chan bool),
	}
}

func (wp *WorkerPool) Start() {
	fmt.Println("Worker Pool запущен")
}

func (wp *WorkerPool) AddWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	workerID := int(atomic.AddInt64(&wp.nextWorkerID, 1))

	worker := &Worker{
		id:       workerID,
		pool:     wp,
		jobQueue: wp.jobQueue,
		quit:     make(chan bool),
	}

	wp.workers[workerID] = worker
	atomic.AddInt64(&wp.workerCount, 1)

	wp.wg.Add(1)
	go worker.start()

	fmt.Printf("Added worker #%d (Total: %d)\n", workerID, atomic.LoadInt64(&wp.workerCount))
}

func (wp *WorkerPool) DeleteWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.workers) == 0 {
		fmt.Println("No workers to remove")
		return
	}

	var workerToRemove *Worker
	var workerID int

	for id, worker := range wp.workers {
		workerToRemove = worker
		workerID = id
		break
	}

	if workerToRemove != nil {
		close(workerToRemove.quit)
		delete(wp.workers, workerID)
		atomic.AddInt64(&wp.workerCount, -1)

		fmt.Printf("Removed worker #%d (Total: %d)\n", workerID, atomic.LoadInt64(&wp.workerCount))
	}
}

func (wp *WorkerPool) AddJob(jobData string) {
	job := Job{
		ID:   int(time.Now().UnixNano()),
		Data: jobData,
	}

	select {
	case wp.jobQueue <- job:
		fmt.Printf("Job added: %s\n", jobData)
	default:
		fmt.Println("Job queue is full!")
	}
}

func (wp *WorkerPool) GetWorkerCount() int64 {
	return atomic.LoadInt64(&wp.workerCount)
}

func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	fmt.Println("Stopping all workers...")

	for _, worker := range wp.workers {
		close(worker.quit)
	}

	close(wp.jobQueue)

	wp.wg.Wait()

	wp.workers = make(map[int]*Worker)
	atomic.StoreInt64(&wp.workerCount, 0)
	atomic.StoreInt64(&wp.nextWorkerID, 0)

	fmt.Println("All workers stopped")
}

func (wp *WorkerPool) GetStatus() {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	fmt.Printf("\nWorker Pool Status:\n")
	fmt.Printf("   Active Workers: %d\n", atomic.LoadInt64(&wp.workerCount))
	fmt.Printf("   Queue Length: %d/%d\n", len(wp.jobQueue), cap(wp.jobQueue))
	fmt.Printf("   Time: %s\n", time.Now().Format("15:04:05"))

	if len(wp.workers) > 0 {
		fmt.Printf("   Worker IDs: ")
		for id := range wp.workers {
			fmt.Printf("#%d ", id)
		}
		fmt.Println()
	}
	fmt.Println()
}

func (w *Worker) start() {
	defer w.pool.wg.Done()

	fmt.Printf("Worker #%d started\n", w.id)

	for {
		select {
		case job, ok := <-w.jobQueue:
			if !ok {
				fmt.Printf("Worker #%d stopped (job queue closed)\n", w.id)
				return
			}
			w.processJob(job)

		case <-w.quit:
			fmt.Printf("Worker #%d stopped (quit signal)\n", w.id)
			return
		}
	}
}

func (w *Worker) processJob(job Job) {
	fmt.Printf("Worker #%d processing job: %s\n", w.id, job.Data)

	time.Sleep(time.Millisecond * 500)

	fmt.Printf("Worker #%d completed job: %s\n", w.id, job.Data)
}
