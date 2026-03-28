; 22: Newton's method for square root
(define (abs x) (if (< x 0) (- x) x))

(define (sqrt-newton x)
  (define (good-enough? guess)
    (< (abs (- (* guess guess) x)) 0.00001))
  (define (improve guess)
    (/ (+ guess (/ x guess)) 2.0))
  (define (sqrt-iter guess)
    (if (good-enough? guess)
        guess
        (sqrt-iter (improve guess))))
  (sqrt-iter 1.0))

(display (sqrt-newton 2.0))   ; => ~1.4142
(newline)
(display (sqrt-newton 9.0))   ; => ~3.0
(newline)
