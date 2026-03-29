; 73: Memoization Decorator
; A generic memoization wrapper using closures and association lists.

(define (memoize f)
  (let ((cache '()))
    (lambda args
      (let ((cached (assoc args cache)))
        (if cached
            (cdr cached)
            (let ((result (apply f args)))
              (set! cache (cons (cons args result) cache))
              result))))))

; Naive recursive fibonacci (exponential without memoization)
(define fib
  (memoize
   (lambda (n)
     (if (< n 2)
         n
         (+ (fib (- n 1)) (fib (- n 2)))))))

; Memoized ackermann (partial - small args only)
(define ack
  (memoize
   (lambda (m n)
     (cond ((= m 0) (+ n 1))
           ((= n 0) (ack (- m 1) 1))
           (else (ack (- m 1) (ack m (- n 1))))))))

(display "Fibonacci (memoized):") (newline)
(let loop ((i 0))
  (when (<= i 20)
    (display "  fib(") (display i) (display ") = ")
    (display (fib i)) (newline)
    (loop (+ i 1))))

(display "Ackermann (memoized):") (newline)
(display "  ack(0,0) = ") (display (ack 0 0)) (newline)
(display "  ack(1,1) = ") (display (ack 1 1)) (newline)
(display "  ack(2,2) = ") (display (ack 2 2)) (newline)
(display "  ack(3,3) = ") (display (ack 3 3)) (newline)
(display "  ack(3,4) = ") (display (ack 3 4)) (newline)
