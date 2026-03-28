; 81: Puzzle 1
; FizzBuzz
(define (fizzbuzz n)
  (let loop ((i 1))
    (when (<= i n)
      (cond
        ((= 0 (modulo i 15)) (display "FizzBuzz"))
        ((= 0 (modulo i 3)) (display "Fizz"))
        ((= 0 (modulo i 5)) (display "Buzz"))
        (else (display i)))
      (display " ")
      (loop (+ i 1)))))
(fizzbuzz 20)
(newline)
