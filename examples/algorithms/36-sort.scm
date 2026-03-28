; 36: Insertion sort
(define (insert x sorted)
  (cond
    ((null? sorted) (list x))
    ((<= x (car sorted)) (cons x sorted))
    (else (cons (car sorted) (insert x (cdr sorted))))))

(define (insertion-sort lst)
  (if (null? lst) '()
      (insert (car lst) (insertion-sort (cdr lst)))))

(display (insertion-sort '(5 3 1 4 2)))  ; => (1 2 3 4 5)
(newline)
