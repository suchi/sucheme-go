; 05: Functions (lambda)
(define (square x) (* x x))
(display (square 5))  ; => 25
(newline)

(define (add a b) (+ a b))
(display (add 3 4))  ; => 7
(newline)

; Anonymous lambda
(display ((lambda (x) (* x 2)) 21))  ; => 42
(newline)
