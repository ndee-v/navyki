package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const minFlagNum = 1

type result struct {
	addr   string
	status int
	err    error
}

func createLogFile(path string) (*os.File, error) {
	return os.Create(fmt.Sprintf("%s%s%s.txt", path, "/results/", time.Now().Format("01-02-15-04-05")))
}

func write(f *os.File, batch []string) (err error) {
	for i := range batch {
		_, err = fmt.Fprintln(f, batch[i])
		if err != nil {
			return err
		}
	}
	return
}

func writer(f *os.File, out <-chan result, batchSize int) (err error) {
	batch := make([]string, 0, batchSize)
	for r := range out {
		if r.err != nil {
			batch = append(batch, fmt.Sprintf("%s %s", r.addr, r.err.Error()))
		} else {
			batch = append(batch, fmt.Sprintf("%s %d", r.addr, r.status))
		}
		if len(batch) == batchSize {
			err = write(f, batch)
			if err != nil {
				return
			}
			batch = batch[:0]
		}
	}
	return write(f, batch)
}

func reader(path string, ch chan<- string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ch <- scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	return nil
}

func sendHTTP(client *http.Client, addr string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func worker(out <-chan string, in chan<- result, retryLimit int, wg *sync.WaitGroup) {
	defer wg.Done()
	client := http.Client{}
	var tryNum int
	var err error
	var resp *http.Response
	for addr := range out {
		tryNum = 0
		for tryNum < retryLimit {
			tryNum++
			resp, err = sendHTTP(&client, addr)
			if err != nil {
				if os.IsTimeout(err) || errors.Is(err, context.DeadlineExceeded) {
					continue
				}
			}
			break
		}
		if resp != nil {
			in <- result{addr: addr, status: resp.StatusCode}
			continue
		}
		in <- result{addr: addr, err: err}
	}
}

func main() {
	dirForResultPath := flag.String("path", ".", "directory for logs")
	inputFilePath := flag.String("input", "./example/urls.txt", "file within URLs")
	workersNum := flag.Int("workers", 5, "number of workers")
	retriesNum := flag.Int("retries", 5, "number of retries")
	batchSize := flag.Int("batch", 3, "batch size")
	flag.Parse()

	if *workersNum < 1 {
		*workersNum = minFlagNum
	}
	if *retriesNum < 1 {
		*retriesNum = minFlagNum
	}
	if *batchSize < 1 {
		*batchSize = minFlagNum
	}
	file, err := createLogFile(*dirForResultPath)
	if err != nil {
		log.Printf("[ERROR]:createLogFile:%s", err.Error())
		return
	}
	defer file.Close()

	resChan := make(chan result)
	inputChan := make(chan string)
	wg := &sync.WaitGroup{}
	wg2 := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		err = reader(*inputFilePath, inputChan)
		if err != nil {
			log.Printf("[ERROR]:reader:%s", err.Error())
		}
		close(inputChan)
		log.Printf("[INFO]:reader: finished")
		wg.Done()
	}()
	go func() {
		err = writer(file, resChan, *batchSize)
		if err != nil {
			log.Printf("[ERROR]:writer:%s", err.Error())
		}
		log.Printf("[INFO]:writer: finished")
		wg.Done()
	}()
	for *workersNum > 0 {
		*workersNum--
		wg2.Add(1)
		go worker(inputChan, resChan, *retriesNum, wg2)
	}
	go func() {
		wg2.Wait()
		close(resChan)
		log.Printf("[INFO]:all workers finished")
	}()
	wg.Wait()
	log.Printf("[INFO]:program finished")
}
