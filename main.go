package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	var workerPool = NewWorkerPool()
	scanner := bufio.NewScanner(os.Stdin)

	workerPool.Start()

	fmt.Println("Worker Pool Management System")
	fmt.Println("Commands: add, delete, job, jobs <N>, status, help, quit")

	for {
		fmt.Print("\n> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := strings.ToLower(parts[0])

		switch command {
		case "add":
			workerPool.AddWorker()

		case "delete", "del":
			workerPool.DeleteWorker()

		case "job":
			fmt.Print("Enter job description: ")
			if scanner.Scan() {
				jobText := scanner.Text()
				if jobText != "" {
					workerPool.AddJob(jobText)
				}
			}

		case "jobs":
			if len(parts) > 1 {
				if count, err := strconv.Atoi(parts[1]); err == nil && count > 0 {
					for i := 1; i <= count; i++ {
						jobText := fmt.Sprintf("Auto job #%d", i)
						workerPool.AddJob(jobText)
					}
				} else {
					fmt.Println("Invalid number")
				}
			} else {
				fmt.Println("Usage: jobs <number>")
			}

		case "status":
			workerPool.GetStatus()

		case "help":
			fmt.Println("\nAvailable commands:")
			fmt.Println("  add        - Add a worker")
			fmt.Println("  delete     - Remove a worker")
			fmt.Println("  job        - Add a single job")
			fmt.Println("  jobs <N>   - Add N jobs")
			fmt.Println("  status     - Show pool status")
			fmt.Println("  help       - Show this help")
			fmt.Println("  quit       - Exit program")

		case "quit", "exit":
			fmt.Println("Shutting down...")
			workerPool.Stop()
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", command)
		}
	}

	workerPool.Stop()
}
