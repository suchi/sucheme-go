; 48: Math example 7
; Triangular number
(define (triangular n) (quotient (* n (+ n 1)) 2))
(display (triangular 10))  ; => 55
(newline)
