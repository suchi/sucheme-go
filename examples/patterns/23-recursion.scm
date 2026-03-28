; 23: Recursion patterns

; List length (recursive)
(define (my-length lst)
  (if (null? lst) 0 (+ 1 (my-length (cdr lst)))))

; List reversal (tail-recursive)
(define (my-reverse lst)
  (let loop ((l lst) (acc '()))
    (if (null? l) acc (loop (cdr l) (cons (car l) acc)))))

; Flatten nested list
(define (flatten lst)
  (cond
    ((null? lst) '())
    ((pair? (car lst))
     (append (flatten (car lst)) (flatten (cdr lst))))
    (else (cons (car lst) (flatten (cdr lst))))))

(display (my-length '(a b c d e)))  ; => 5
(newline)
(display (my-reverse '(1 2 3)))  ; => (3 2 1)
(newline)
(display (flatten '(1 (2 (3 4) 5) 6)))  ; => (1 2 3 4 5 6)
(newline)
