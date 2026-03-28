; 87: Puzzle 7
; All permutations
(define (permutations lst)
  (if (null? lst) (list (list))
      (apply append
        (map (lambda (x)
               (map (lambda (p) (cons x p))
                    (permutations (remove x lst))))
             lst))))
(define (remove x lst)
  (filter (lambda (y) (not (equal? x y))) lst))
(display (length (permutations (list 1 2 3 4))))  ; => 24
(newline)
