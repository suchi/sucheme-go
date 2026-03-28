; 46: Math example 5
; Fibonacci iterative
(define (fib-iter n)
  (let loop ((i 0) (a 0) (b 1))
    (if (= i n) a (loop (+ i 1) b (+ a b)))))
(display (fib-iter 50))
(newline)
