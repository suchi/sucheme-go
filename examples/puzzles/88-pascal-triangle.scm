; 88: Puzzle 8
; Pascal triangle row
(define (pascal-row n)
  (let loop ((row (list 1)) (i 0))
    (if (= i n) row
        (loop (let next ((prev row) (acc (list 1)))
                (if (null? (cdr prev)) (reverse (cons 1 acc))
                    (next (cdr prev) (cons (+ (car prev) (car (cdr prev))) acc))))
              (+ i 1)))))
(display (pascal-row 5))
(newline)
