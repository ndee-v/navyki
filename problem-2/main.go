package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type result struct {
	addr   string
	status int
}

func createLogFile(path string) (*os.File, error) {
	return os.Create(fmt.Sprintf("%s%s%s.txt", path, "/results/", time.Now().Format("01-02-15-04-05")))
}

func writer(f *os.File, source <-chan result) error {

	return nil
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

func main() {
	dirForResultPath := flag.String("path", ".", "directory for logs")
	inputFilePath := flag.String("input", "./example/urls.txt", "file within URLs")
	workersNum := flag.Uint("workers", 5, "number of workers")
	fmt.Println("num of workers is: ", *workersNum)
	flag.Parse()
	file, err := createLogFile(*dirForResultPath)
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	resChan := make(chan result)
	inputChan := make(chan string)
	wg := &sync.WaitGroup{}
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
		err = writer(file, resChan)
		if err != nil {
			log.Printf("[ERROR]:writer:%s", err.Error())
		}
		close(resChan)
		log.Printf("[INFO]:writer: finished")
		wg.Done()
	}()
	for url := range inputChan {
		fmt.Println(url)
	}
	wg.Wait()
	log.Printf("[INFO]:finished")
}
