; 86: Puzzle 6
; Tower of Hanoi (count moves)
(define (hanoi-count n) (- (fast-expt 2 n) 1))
(define (fast-expt b e)
  (cond ((= e 0) 1) ((= (modulo e 2) 0) (let ((h (fast-expt b (quotient e 2)))) (* h h)))
        (else (* b (fast-expt b (- e 1))))))
(display (hanoi-count 10))  ; => 1023
(newline)
