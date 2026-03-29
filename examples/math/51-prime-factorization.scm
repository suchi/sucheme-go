; 51: Prime Factorization
; Decompose a number into its prime factors using trial division.

(define (prime-factors n)
  (define (factor n d)
    (cond ((<= n 1) '())
          ((= 0 (modulo n d))
           (cons d (factor (/ n d) d)))
          (else (factor n (+ d 1)))))
  (factor n 2))

(display "Prime factors of 360: ")
(display (prime-factors 360))
(newline)

(display "Prime factors of 97 (prime): ")
(display (prime-factors 97))
(newline)

(display "Prime factors of 1024: ")
(display (prime-factors 1024))
(newline)
