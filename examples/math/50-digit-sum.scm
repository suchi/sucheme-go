; 50: Math example 9
; Digit sum
(define (digit-sum n)
  (if (= n 0) 0 (+ (modulo n 10) (digit-sum (quotient n 10)))))
(display (digit-sum 12345))  ; => 15
(newline)
