; 11: Map, filter, for-each
(display (map (lambda (x) (* x x)) '(1 2 3 4 5)))
(newline)  ; => (1 4 9 16 25)

(display (filter (lambda (x) (> x 3)) '(1 2 3 4 5)))
(newline)  ; => (4 5)

(for-each (lambda (x) (display x) (display " ")) '(a b c))
(newline)  ; => a b c
