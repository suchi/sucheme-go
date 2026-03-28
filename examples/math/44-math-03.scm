; 44: Math example 3
(define (sum-list lst)
  (let loop ((l lst) (acc 0))
    (if (null? l) acc (loop (cdr l) (+ acc (car l))))))
(display (sum-list (list 1 2 3 4 5)))  ; => 15
(newline)
