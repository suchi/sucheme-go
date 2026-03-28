; 30: Collatz conjecture
(define (collatz n)
  (let loop ((n n) (steps 0))
    (cond
      ((= n 1) steps)
      ((= (modulo n 2) 0) (loop (/ n 2) (+ steps 1)))
      (else (loop (+ (* 3 n) 1) (+ steps 1))))))

(display (collatz 27))  ; => 111
(newline)
