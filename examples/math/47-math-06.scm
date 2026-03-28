; 47: Math example 6
; Power of 2 check
(define (pow2? n)
  (cond
    ((<= n 0) #f)
    ((= n 1) #t)
    ((= (modulo n 2) 1) #f)
    (else (pow2? (quotient n 2)))))

(display (pow2? 16))  ; => #t
(newline)
(display (pow2? 15))  ; => #f
(newline)
(display (pow2? 1))   ; => #t
(newline)
