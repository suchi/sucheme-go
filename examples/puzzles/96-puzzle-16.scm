; 96: Combination Generator
; Generate all k-element subsets of a list.
; C(n,k) = C(n-1,k-1) + C(n-1,k)

(define (combinations k lst)
  (cond ((= k 0) '(()))
        ((null? lst) '())
        (else
         (append
          ; Combinations including the first element
          (map (lambda (c) (cons (car lst) c))
               (combinations (- k 1) (cdr lst)))
          ; Combinations excluding the first element
          (combinations k (cdr lst))))))

(display "C(5,2) - Choose 2 from (1 2 3 4 5):") (newline)
(for-each (lambda (c) (display "  ") (display c) (newline))
          (combinations 2 '(1 2 3 4 5)))

(display "Count: ") (display (length (combinations 2 '(1 2 3 4 5)))) (newline)

(newline)
(display "C(5,3) - Choose 3 from (a b c d e):") (newline)
(for-each (lambda (c) (display "  ") (display c) (newline))
          (combinations 3 '(a b c d e)))

(display "Count: ") (display (length (combinations 3 '(a b c d e)))) (newline)

(newline)
; Edge cases
(display "C(4,0): ") (display (combinations 0 '(1 2 3 4))) (newline)
(display "C(4,4): ") (display (combinations 4 '(1 2 3 4))) (newline)
(display "C(3,5): ") (display (combinations 5 '(1 2 3))) (newline)
