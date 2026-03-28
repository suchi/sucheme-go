; 09: Vectors
(define v (vector 10 20 30))
(display (vector-ref v 1))  ; => 20
(newline)
(vector-set! v 0 99)
(display v)  ; => #(99 20 30)
(newline)
(display (vector->list v))  ; => (99 20 30)
(newline)
