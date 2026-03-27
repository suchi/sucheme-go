; Factorial in Sucheme
(define (factorial n)
  (if (= n 0)
      1
      (* n (factorial (- n 1)))))

(display "10! = ")
(display (factorial 10))
(newline)

; Tail-recursive version
(define (factorial-iter n acc)
  (if (= n 0)
      acc
      (factorial-iter (- n 1) (* n acc))))

(display "20! = ")
(display (factorial-iter 20 1))
(newline)
