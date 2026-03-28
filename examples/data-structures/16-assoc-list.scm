; 16: Association lists (key-value pairs)
(define env '((x . 1) (y . 2) (z . 3)))

(define (lookup key alist)
  (cond
    ((null? alist) #f)
    ((equal? key (car (car alist))) (cdr (car alist)))
    (else (lookup key (cdr alist)))))

(display (lookup 'y env))  ; => 2
(newline)
(display (lookup 'w env))  ; => #f
(newline)
