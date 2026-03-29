; 100: Sudoku Solver (4x4)
; Solve a 4x4 Sudoku puzzle using backtracking search with validity checks.
; 0 represents an empty cell. Values are 1-4.

(define (list-set lst idx val)
  (let loop ((l lst) (i 0))
    (if (null? l) '()
        (cons (if (= i idx) val (car l))
              (loop (cdr l) (+ i 1))))))

(define (grid-ref grid r c) (list-ref (list-ref grid r) c))
(define (grid-set grid r c val)
  (list-set grid r (list-set (list-ref grid r) c val)))

(define (valid-placement? grid r c val)
  ; Check row
  (define (check-row)
    (let loop ((j 0))
      (cond ((= j 4) #t)
            ((= (grid-ref grid r j) val) #f)
            (else (loop (+ j 1))))))
  ; Check column
  (define (check-col)
    (let loop ((i 0))
      (cond ((= i 4) #t)
            ((= (grid-ref grid i c) val) #f)
            (else (loop (+ i 1))))))
  ; Check 2x2 box
  (define (check-box)
    (let ((br (* 2 (quotient r 2)))
          (bc (* 2 (quotient c 2))))
      (let loop-r ((i br))
        (cond ((= i (+ br 2)) #t)
              (else (let loop-c ((j bc))
                      (cond ((= j (+ bc 2)) (loop-r (+ i 1)))
                            ((= (grid-ref grid i j) val) #f)
                            (else (loop-c (+ j 1))))))))))
  (and (check-row) (check-col) (check-box)))

; Find first empty cell (value = 0)
(define (find-empty grid)
  (let loop-r ((r 0))
    (if (= r 4) #f
        (let loop-c ((c 0))
          (cond ((= c 4) (loop-r (+ r 1)))
                ((= (grid-ref grid r c) 0) (cons r c))
                (else (loop-c (+ c 1))))))))

(define (solve grid)
  (let ((empty (find-empty grid)))
    (if (not empty)
        grid  ; solved!
        (let ((r (car empty)) (c (cdr empty)))
          (let try ((val 1))
            (cond ((> val 4) #f)  ; no valid value
                  ((valid-placement? grid r c val)
                   (let ((result (solve (grid-set grid r c val))))
                     (if result result (try (+ val 1)))))
                  (else (try (+ val 1)))))))))

(define (print-grid grid)
  (for-each (lambda (row)
              (display "  ")
              (for-each (lambda (v)
                          (display v) (display " "))
                        row)
              (newline))
            grid))

; Puzzle 1
(define puzzle1
  '((0 0 3 0)
    (3 0 0 2)
    (2 0 0 1)
    (0 1 0 0)))

(display "Puzzle 1:") (newline)
(print-grid puzzle1)
(display "Solution:") (newline)
(let ((sol (solve puzzle1)))
  (if sol (print-grid sol) (display "  No solution!" )))
(newline)

; Puzzle 2
(define puzzle2
  '((0 2 0 0)
    (0 0 0 3)
    (4 0 0 0)
    (0 0 1 0)))

(display "Puzzle 2:") (newline)
(print-grid puzzle2)
(display "Solution:") (newline)
(let ((sol (solve puzzle2)))
  (if sol (print-grid sol) (display "  No solution!")))
(newline)
