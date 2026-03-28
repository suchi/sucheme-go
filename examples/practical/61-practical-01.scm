; 61: Practical example 1
; Count occurrences in list
(define (count-if pred lst)
  (let loop ((l lst) (n 0))
    (cond ((null? l) n)
          ((pred (car l)) (loop (cdr l) (+ n 1)))
          (else (loop (cdr l) n)))))
(display (count-if even? (list 1 2 3 4 5 6)))  ; count of even numbers
(newline)
