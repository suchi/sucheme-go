; 14: Currying
(define (curry f)
  (lambda (a) (lambda (b) (f a b))))

(define add (curry +))
(define add5 (add 5))
(display (add5 3))  ; => 8
(newline)

(define mul (curry *))
(define double (mul 2))
(display (double 21))  ; => 42
(newline)
