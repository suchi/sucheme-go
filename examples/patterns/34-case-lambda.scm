; 34: case-lambda
(define add
  (case-lambda
    (() 0)
    ((x) x)
    ((x y) (+ x y))
    ((x y z) (+ x y z))))

(display (add))      ; => 0
(newline)
(display (add 1))    ; => 1
(newline)
(display (add 1 2))  ; => 3
(newline)
(display (add 1 2 3)) ; => 6
(newline)
