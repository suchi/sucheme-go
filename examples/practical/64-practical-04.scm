; 64: Practical example 4
; Take first n elements
(define (take n lst)
  (if (or (= n 0) (null? lst)) (list)
      (cons (car lst) (take (- n 1) (cdr lst)))))
(display (take 3 (list 1 2 3 4 5)))  ; => (1 2 3)
(newline)
