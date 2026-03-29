; 66: Practical example 6
; Last element of list
(define (last lst)
  (if (null? (cdr lst)) (car lst) (last (cdr lst))))
(display (last (list 1 2 3)))  ; => 3
(newline)
