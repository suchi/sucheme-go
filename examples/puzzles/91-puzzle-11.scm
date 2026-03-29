; 91: N-Queens Problem
; Place N queens on an NxN chessboard so that no two queens attack each other.
; Uses backtracking with tail-call optimization.

(define (queens n)
  (define (safe? col row placed)
    (let loop ((c 0) (rest placed))
      (if (null? rest)
          #t
          (let ((r (car rest)))
            (if (or (= r row)
                    (= (abs (- col c)) (abs (- row r))))
                #f
                (loop (+ c 1) (cdr rest)))))))
  (define (abs x) (if (< x 0) (- x) x))
  (define (solve col placed)
    (if (= col n)
        (list (reverse placed))
        (let loop ((row 0) (solutions '()))
          (if (= row n)
              solutions
              (loop (+ row 1)
                    (if (safe? col row placed)
                        (append solutions (solve (+ col 1) (cons row placed)))
                        solutions))))))
  (solve 0 '()))

(display "4-Queens solutions:") (newline)
(for-each (lambda (sol)
            (display "  ") (display sol) (newline))
          (queens 4))

(display "Number of 5-Queens solutions: ")
(display (length (queens 5)))
(newline)

(display "Number of 6-Queens solutions: ")
(display (length (queens 6)))
(newline)

; Show first 8-queens solution as a board
(display "First 8-Queens solution:") (newline)
(let ((sol (car (queens 8))))
  (let loop-r ((r 0))
    (when (< r 8)
      (display "  ")
      (let loop-c ((c 0))
        (when (< c 8)
          (display (if (= (list-ref sol c) r) "Q " ". "))
          (loop-c (+ c 1))))
      (newline)
      (loop-r (+ r 1)))))
