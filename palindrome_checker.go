package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

// Checks if a number is a palindrome.
func isPalindrome(n int) bool {
	s := strconv.Itoa(n)
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}
	return true
}

// Checks if a number is prime.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// Worker that checks if numbers are prime palindromes and sends valid ones to results.
func worker(ctx context.Context, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-jobs:
			if !ok {
				return
			}
			if isPrime(n) && isPalindrome(n) {
				results <- n
			}
		}
	}
}

// Generator sends numbers to the jobs channel until the context is canceled.
func generator(ctx context.Context, jobs chan<- int) {
	defer close(jobs)
	for i := 2; ; i++ {
		select {
		case <-ctx.Done():
			return
		case jobs <- i:
		}
	}
}

func palindromeChecker() {
	var N int
	fmt.Print("Enter N: ")
	_, err := fmt.Scanln(&N)
	if err != nil {
		return
	}

	if N < 1 || N > 50 {
		fmt.Println("N must be between 1 and 50.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int, 100)
	results := make(chan int, N)

	var wg sync.WaitGroup

	workerCount := 4

	// Start worker Goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg)
	}

	// Start number generator
	go generator(ctx, jobs)

	var primes []int
	sum := 0

	// Collect N prime palindromes
	for len(primes) < N {
		p := <-results
		primes = append(primes, p)
		sum += p
	}

	cancel()  // Stop generator and workers
	wg.Wait() // Wait for workers to exit
	close(results)

	// Output results
	fmt.Println("Prime Palindromic Numbers:")
	for _, p := range primes {
		fmt.Print(p, " ")
	}
	fmt.Printf("\nSum: %d\n", sum)
}
