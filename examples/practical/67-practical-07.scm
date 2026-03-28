; 67: Practical example 7
; Remove duplicates
(define (remove-duplicates lst)
  (let loop ((l lst) (acc (list)))
    (if (null? l) (reverse acc)
        (if (member (car l) acc) (loop (cdr l) acc)
            (loop (cdr l) (cons (car l) acc))))))

(define (member x lst)
  (cond ((null? lst) #f) ((equal? x (car lst)) lst) (else (member x (cdr lst)))))

(display (remove-duplicates (list 1 2 3 2 1 4)))  ; => (1 2 3 4)
(newline)
