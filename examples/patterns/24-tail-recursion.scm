; 24: Tail recursion (TCO)
; These can run with millions of iterations without stack overflow

(define (sum-to n)
  (let loop ((i 0) (acc 0))
    (if (> i n) acc (loop (+ i 1) (+ acc i)))))

(display (sum-to 100000))  ; => 5000050000
(newline)

; Mutual tail recursion
(define (my-even? n)
  (if (= n 0) #t (my-odd? (- n 1))))
(define (my-odd? n)
  (if (= n 0) #f (my-even? (- n 1))))

(display (my-even? 10000))  ; => #t
(newline)
