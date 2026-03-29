; 72: Streams (Lazy Evaluation) using Closures
; Implement infinite streams with delay/force semantics via closures.

(define (stream-cons head tail-thunk)
  (cons head tail-thunk))

(define (stream-car s) (car s))

(define (stream-cdr s)
  ((cdr s)))

(define (stream-take n s)
  (if (or (= n 0) (null? s))
      '()
      (cons (stream-car s)
            (stream-take (- n 1) (stream-cdr s)))))

(define (stream-map f s)
  (stream-cons (f (stream-car s))
               (lambda () (stream-map f (stream-cdr s)))))

(define (stream-filter pred s)
  (if (pred (stream-car s))
      (stream-cons (stream-car s)
                   (lambda () (stream-filter pred (stream-cdr s))))
      (stream-filter pred (stream-cdr s))))

; Infinite stream of natural numbers starting from n
(define (naturals-from n)
  (stream-cons n (lambda () (naturals-from (+ n 1)))))

; Fibonacci stream
(define (fib-stream a b)
  (stream-cons a (lambda () (fib-stream b (+ a b)))))

(define nats (naturals-from 1))
(display "First 10 naturals: ")
(display (stream-take 10 nats))
(newline)

(display "First 10 squares: ")
(display (stream-take 10 (stream-map (lambda (x) (* x x)) nats)))
(newline)

(display "First 10 even numbers: ")
(display (stream-take 10 (stream-filter even? nats)))
(newline)

(display "First 15 Fibonacci: ")
(display (stream-take 15 (fib-stream 0 1)))
(newline)
