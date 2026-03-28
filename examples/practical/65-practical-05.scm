; 65: Practical example 5
; Drop first n elements
(define (drop n lst)
  (if (or (= n 0) (null? lst)) lst
      (drop (- n 1) (cdr lst))))
(display (drop 2 (list 1 2 3 4 5)))  ; => (3 4 5)
(newline)
