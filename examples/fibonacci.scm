; Fibonacci in Sucheme
(define (fib n)
  (if (< n 2)
      n
      (+ (fib (- n 1)) (fib (- n 2)))))

(display "fib(10) = ")
(display (fib 10))
(newline)

; Tail-recursive version
(define (fib-iter n)
  (let loop ((i 0) (a 0) (b 1))
    (if (= i n)
        a
        (loop (+ i 1) b (+ a b)))))

(display "fib(30) = ")
(display (fib-iter 30))
(newline)
