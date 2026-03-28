; 04: Conditionals
(display (if #t "yes" "no"))  ; => yes
(newline)

(display (cond
  ((> 3 5) "a")
  ((< 3 5) "b")
  (else "c")))  ; => b
(newline)

(display (and 1 2 3))  ; => 3
(newline)
(display (or #f #f 42))  ; => 42
(newline)
