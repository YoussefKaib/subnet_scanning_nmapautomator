package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

var (
	path       = "./"
	subnet     *string
	numWorkers *int
	scanType   *string
)

func init() {
	scanType = flag.String("scanType", "Basic", "a string")
	flag.Parse()
}

// Here's the worker, of which we'll run several
// concurrent instances. These workers will receive
// work on the `jobs` channel and send the corresponding
// results on `results`. We'll sleep a second per job to
// simulate an expensive task.
func worker(ip string, id int, jobs <-chan int, wg *sync.WaitGroup) {
	for j := range jobs {
		fmt.Println("worker", id, "started  job", j)
		command := `./bin/nmapautomator.sh ` + ip + ` ` + *scanType
		_, err := exec.Command("/bin/sh", "-c", command).Output()
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println("worker", id, "finished job", j)
		wg.Done()
	}
}

func main() {

	var wg sync.WaitGroup

	err := os.Chdir("./")
	if err != nil {
		log.Fatal("failed change to root project directory error: ", err)
	}

	ips := ipList()

	// In order to use our pool of workers we need to send
	// them work and collect their results. We make 2
	// channels for this.
	jobs := make(chan int, len(ips))

	// This starts up 3 workers, initially blocked
	// because there are no jobs yet.
	for _, ip := range ips {
		wg.Add(1)
		go worker(ip, len(ips), jobs, &wg)

	}

	// Here we send 5 `jobs` and then `close` that
	// channel to indicate that's all the work we have.
	for j := 1; j <= len(ips); j++ {
		jobs <- j
	}
	close(jobs)

	// Finally we collect all the results of the work.
	// This also ensures that the worker goroutines have
	// finished.
	wg.Wait()
}

func ipList() []string {
	listIPs := []string{}
	file, _ := os.Open(path + "targets.txt")

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		listIPs = append(listIPs, scanner.Text())
	}

	return listIPs
}
