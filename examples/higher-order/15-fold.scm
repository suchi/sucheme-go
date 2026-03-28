; 15: Fold (reduce)
(define (fold-left f init lst)
  (if (null? lst)
      init
      (fold-left f (f init (car lst)) (cdr lst))))

(display (fold-left + 0 '(1 2 3 4 5)))  ; => 15
(newline)

(display (fold-left * 1 '(1 2 3 4 5)))  ; => 120
(newline)

; fold-right
(define (fold-right f init lst)
  (if (null? lst)
      init
      (f (car lst) (fold-right f init (cdr lst)))))

(display (fold-right cons '() '(1 2 3)))  ; => (1 2 3)
(newline)
