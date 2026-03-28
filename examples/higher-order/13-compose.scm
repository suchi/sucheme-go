; 13: Function composition
(define (compose f g)
  (lambda (x) (f (g x))))

(define inc (lambda (x) (+ x 1)))
(define double (lambda (x) (* x 2)))

(define inc-then-double (compose double inc))
(display (inc-then-double 3))  ; => 8
(newline)

(define double-then-inc (compose inc double))
(display (double-then-inc 3))  ; => 7
(newline)
