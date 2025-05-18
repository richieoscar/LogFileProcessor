package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	workerCount = 4 // Number of concurrent workers
)

// KeywordCount holds the keyword and its frequency.
type KeywordCount struct {
	Keyword string
	Count   int
}

// readLines reads lines from a file and sends them to a channel in batches.
func ReadLines(filePath string, lineChan chan<- []string, batchSize int, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		close(lineChan)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var batch []string

	for scanner.Scan() {
		batch = append(batch, scanner.Text())
		if len(batch) >= batchSize {
			lineChan <- batch
			batch = nil
		}
	}
	if len(batch) > 0 {
		lineChan <- batch
	}
	close(lineChan)
}

// worker counts keyword occurrences in lines and sends results via a channel.
func Worker(keywords []string, lineChan <-chan []string, countChan chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	for lines := range lineChan {
		counts := make(map[string]int)
		for _, line := range lines {
			for _, keyword := range keywords {
				if strings.Contains(line, keyword) {
					counts[keyword]++
				}
			}
		}
		countChan <- counts
	}
}

// mergeCounts aggregates results from all workers.
func MergeCounts(countChan <-chan map[string]int, finalCounts map[string]int, done chan<- bool) {
	for partial := range countChan {
		for keyword, count := range partial {
			finalCounts[keyword] += count
		}
	}
	done <- true
}

// sortCounts returns sorted keyword counts.
func SortCounts(counts map[string]int) []KeywordCount {
	var result []KeywordCount
	for k, v := range counts {
		result = append(result, KeywordCount{k, v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

func process_log(filePath string) {
	keywords := []string{"ERROR", "WARN", "INFO"} // Define your keywords here
	batchSize := 100                              // Lines per batch

	lineChan := make(chan []string, workerCount)
	countChan := make(chan map[string]int, workerCount)
	finalCounts := make(map[string]int)
	done := make(chan bool)

	var readerWg sync.WaitGroup
	var workerWg sync.WaitGroup

	readerWg.Add(1)
	go ReadLines(filePath, lineChan, batchSize, &readerWg)

	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go Worker(keywords, lineChan, countChan, &workerWg)
	}

	go func() {
		readerWg.Wait()
		workerWg.Wait()
		close(countChan)
	}()

	go MergeCounts(countChan, finalCounts, done)

	<-done // Wait until merging is complete

	// Print sorted results
	sorted := SortCounts(finalCounts)
	for _, kc := range sorted {
		fmt.Printf("%s: %d\n", kc.Keyword, kc.Count)
	}

}
