; 37: Merge sort
(define (merge a b)
  (cond
    ((null? a) b)
    ((null? b) a)
    ((<= (car a) (car b)) (cons (car a) (merge (cdr a) b)))
    (else (cons (car b) (merge a (cdr b))))))

(define (split lst)
  (let loop ((l lst) (a '()) (b '()) (toggle #t))
    (if (null? l)
        (list (reverse a) (reverse b))
        (if toggle
            (loop (cdr l) (cons (car l) a) b #f)
            (loop (cdr l) a (cons (car l) b) #t)))))

(define (mergesort lst)
  (if (or (null? lst) (null? (cdr lst))) lst
      (let ((parts (split lst)))
        (merge (mergesort (car parts))
               (mergesort (car (cdr parts)))))))

(display (mergesort '(38 27 43 3 9 82 10)))
(newline)  ; => (3 9 10 27 38 43 82)
