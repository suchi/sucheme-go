; 53: Catalan Numbers
; The n-th Catalan number counts many combinatorial structures:
; balanced parentheses, binary trees, non-crossing partitions, etc.
; C(n) = C(2n,n) / (n+1) = (2n)! / ((n+1)! * n!)

(define (factorial n)
  (let loop ((i 1) (acc 1))
    (if (> i n) acc (loop (+ i 1) (* acc i)))))

(define (catalan n)
  (/ (factorial (* 2 n))
     (* (factorial (+ n 1)) (factorial n))))

; Recursive relation: C(n) = sum of C(i)*C(n-1-i) for i=0..n-1
(define (catalan-recursive n)
  (if (<= n 1)
      1
      (let loop ((i 0) (sum 0))
        (if (= i n)
            sum
            (loop (+ i 1)
                  (+ sum (* (catalan-recursive i)
                            (catalan-recursive (- n 1 i)))))))))

(display "Catalan numbers (formula):") (newline)
(let loop ((i 0))
  (when (<= i 10)
    (display "  C(") (display i) (display ") = ")
    (display (catalan i)) (newline)
    (loop (+ i 1))))

(display "Verification via recursive relation (first 8):") (newline)
(let loop ((i 0))
  (when (<= i 7)
    (display "  C(") (display i) (display ") = ")
    (display (catalan-recursive i)) (newline)
    (loop (+ i 1))))
