package evaluator

import (
	"testing"

	"github.com/suchi/sucheme-go/pkg/builtin"
	"github.com/suchi/sucheme-go/pkg/env"
)

func setupBench() *env.Env {
	e := env.Global()
	builtin.RegisterBuiltins(e)
	RegisterApply(e)
	return e
}

// BenchmarkTakeuchi benchmarks the Takeuchi function (tak 18 12 6).
func BenchmarkTakeuchi(b *testing.B) {
	e := setupBench()
	_, err := EvalString(`
		(define (tak x y z)
			(if (not (< y x))
				z
				(tak (tak (- x 1) y z)
					 (tak (- y 1) z x)
					 (tak (- z 1) x y))))
	`, e)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EvalString("(tak 18 12 6)", e)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFibonacci benchmarks naive recursive fibonacci(20).
func BenchmarkFibonacci(b *testing.B) {
	e := setupBench()
	_, err := EvalString(`
		(define (fib n)
			(if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))
	`, e)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EvalString("(fib 20)", e)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTailRecursion benchmarks deep tail-recursive loop.
func BenchmarkTailRecursion(b *testing.B) {
	e := setupBench()
	_, err := EvalString(`
		(define (loop n)
			(if (= n 0) 0 (loop (- n 1))))
	`, e)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EvalString("(loop 100000)", e)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListOperations benchmarks list creation and reversal.
func BenchmarkListOperations(b *testing.B) {
	e := setupBench()
	_, err := EvalString(`
		(define (make-list n)
			(let loop ((i 0) (acc '()))
				(if (= i n) acc (loop (+ i 1) (cons i acc)))))
	`, e)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EvalString("(reverse (make-list 1000))", e)
		if err != nil {
			b.Fatal(err)
		}
	}
}
