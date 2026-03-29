; 85: Puzzle 5
; Reverse integer
(define (reverse-int n)
  (let loop ((n n) (acc 0))
    (if (= n 0) acc
        (loop (quotient n 10) (+ (* acc 10) (modulo n 10))))))
(display (reverse-int 12345))  ; => 54321
(newline)
