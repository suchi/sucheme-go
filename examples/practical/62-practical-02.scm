; 62: Practical example 2
; Zip two lists
(define (zip a b)
  (if (or (null? a) (null? b)) (list)
      (cons (list (car a) (car b)) (zip (cdr a) (cdr b)))))
(display (zip (list 1 2 3) (list "a" "b" "c")))
(newline)
