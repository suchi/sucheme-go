; 41: Accumulate (generic fold)
(define (accumulate op init lst)
  (if (null? lst) init
      (op (car lst) (accumulate op init (cdr lst)))))

(display (accumulate + 0 '(1 2 3 4 5)))   ; => 15
(newline)
(display (accumulate * 1 '(1 2 3 4 5)))   ; => 120
(newline)
(display (accumulate cons '() '(a b c)))   ; => (a b c)
(newline)
