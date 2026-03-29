; 97: Longest Common Subsequence (LCS)
; Find the longest subsequence common to two sequences.
; lcs-length: bottom-up DP for computing length efficiently.
; lcs: naive recursive reconstruction (exponential, for small inputs only).

(define (lcs-length xs ys)
  (let ((m (length xs)) (n (length ys)))
    ; Build DP table bottom-up
    (define (build-row prev-row y-elem xs-list)
      (let loop ((xs xs-list) (j 0) (prev-left 0) (row '()))
        (if (null? xs)
            (reverse row)
            (let ((diag (list-ref prev-row j))
                  (up (list-ref prev-row (+ j 1)))
                  (val (if (equal? (car xs) y-elem)
                           (+ diag 1)
                           (max up prev-left))))
              (loop (cdr xs) (+ j 1) val (cons val row))))))
    (define (max a b) (if (> a b) a b))
    ; Initial row of zeros (length m+1)
    (define (zeros n)
      (if (= n 0) '() (cons 0 (zeros (- n 1)))))
    (let loop ((ys-rest ys) (prev (zeros (+ m 1))))
      (if (null? ys-rest)
          (list-ref prev m)
          (let ((new-row (cons 0 (build-row prev (car ys-rest) xs))))
            (loop (cdr ys-rest) new-row))))))

; Reconstruct LCS using recursive approach
(define (lcs xs ys)
  (cond ((or (null? xs) (null? ys)) '())
        ((equal? (car xs) (car ys))
         (cons (car xs) (lcs (cdr xs) (cdr ys))))
        (else
         (let ((a (lcs (cdr xs) ys))
               (b (lcs xs (cdr ys))))
           (if (> (length a) (length b)) a b)))))

(display "LCS of (A B C D E) and (A C E):") (newline)
(display "  LCS: ") (display (lcs '(A B C D E) '(A C E))) (newline)
(display "  Length: ") (display (lcs-length '(A B C D E) '(A C E))) (newline)

(display "LCS of (1 2 3 4 5) and (2 4 6):") (newline)
(display "  LCS: ") (display (lcs '(1 2 3 4 5) '(2 4 6))) (newline)
(display "  Length: ") (display (lcs-length '(1 2 3 4 5) '(2 4 6))) (newline)

(display "LCS of (A B C) and (D E F):") (newline)
(display "  LCS: ") (display (lcs '(A B C) '(D E F))) (newline)
(display "  Length: ") (display (lcs-length '(A B C) '(D E F))) (newline)

(display "LCS of (G A C) and (A G C A T):") (newline)
(display "  LCS: ") (display (lcs '(G A C) '(A G C A T))) (newline)
(display "  Length: ") (display (lcs-length '(G A C) '(A G C A T))) (newline)
