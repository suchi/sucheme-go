; 52: Combinatorics - Binomial Coefficients (C(n,k))
; Compute C(n,k) using the multiplicative formula to avoid large intermediates.

(define (binomial n k)
  (if (or (< k 0) (> k n))
      0
      (let ((k (min k (- n k))))
        (let loop ((i 0) (result 1))
          (if (= i k)
              result
              (loop (+ i 1)
                    (/ (* result (- n i)) (+ i 1))))))))

(display "C(10,3) = ") (display (binomial 10 3)) (newline)
(display "C(20,10) = ") (display (binomial 20 10)) (newline)
(display "C(5,0) = ") (display (binomial 5 0)) (newline)
(display "C(5,5) = ") (display (binomial 5 5)) (newline)
(display "C(0,0) = ") (display (binomial 0 0)) (newline)

; Pascal's triangle (first 8 rows)
(display "Pascal's triangle:") (newline)
(let loop ((n 0))
  (when (<= n 7)
    (let inner ((k 0))
      (when (<= k n)
        (display (binomial n k))
        (display " ")
        (inner (+ k 1))))
    (newline)
    (loop (+ n 1))))
