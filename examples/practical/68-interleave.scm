; 68: Practical example 8
; Interleave two lists
(define (interleave a b)
  (cond ((null? a) b) ((null? b) a)
        (else (cons (car a) (cons (car b) (interleave (cdr a) (cdr b)))))))
(display (interleave (list 1 3 5) (list 2 4 6)))  ; => (1 2 3 4 5 6)
(newline)
