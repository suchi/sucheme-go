; 45: Math example 4
(define (average lst)
  (let loop ((l lst) (sum 0) (count 0))
    (if (null? l) (/ sum count)
        (loop (cdr l) (+ sum (car l)) (+ count 1)))))
(display (average (list 10 20 30)))  ; => 20
(newline)
