; 43: Math example 2
(define (max-of-list lst)
  (if (null? (cdr lst)) (car lst)
      (let ((rest-max (max-of-list (cdr lst))))
        (if (> (car lst) rest-max) (car lst) rest-max))))
(display (max-of-list (list 3 1 4 1 5 9)))  ; => 9
(newline)
