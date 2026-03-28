; 47: Math example 6
; Power of 2
(define (power-of-2? n)
  (and (> n 0) (= 0 (modulo n 2))
       (or (= n 1) (power-of-2? (/ n 2)))))
; Note: simplified
(define (pow2? n)
  (cond ((= n 1) #t) ((= (modulo n 2) 1) #f) (else (pow2? (quotient n 2)))))
(display (pow2? 16))  ; => #t
(newline)
(display (pow2? 15))  ; => #f
(newline)
