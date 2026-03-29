; 94: 0/1 Knapsack Problem (Dynamic Programming)
; Given items with weights and values, find the maximum value that fits in capacity.
; Uses bottom-up DP with a 2D table represented as nested lists.

(define (make-row size init)
  (let loop ((i 0))
    (if (= i size) '() (cons init (loop (+ i 1))))))

(define (knapsack weights values capacity)
  (let ((n (length weights)))
    ; Build DP table: dp[i][w] = max value using items 0..i-1 with capacity w
    (let loop-i ((i 0) (prev (make-row (+ capacity 1) 0)))
      (if (= i n)
          (list-ref prev capacity)
          (let ((wi (list-ref weights i))
                (vi (list-ref values i)))
            (let loop-w ((w 0) (row '()))
              (if (> w capacity)
                  (loop-i (+ i 1) (reverse row))
                  (let ((without (list-ref prev w))
                        (with (if (>= w wi)
                                  (+ vi (list-ref prev (- w wi)))
                                  0)))
                    (loop-w (+ w 1)
                            (cons (if (> with without) with without) row))))))))))

; Example 1
(let ((weights '(2 3 4 5))
      (values  '(3 4 5 6))
      (capacity 8))
  (display "Knapsack Problem 1:") (newline)
  (display "  Items (weight, value): ")
  (let loop ((w weights) (v values))
    (when (pair? w)
      (display "(") (display (car w)) (display ",") (display (car v)) (display ") ")
      (loop (cdr w) (cdr v))))
  (newline)
  (display "  Capacity: ") (display capacity) (newline)
  (display "  Max value: ") (display (knapsack weights values capacity)) (newline))

; Example 2
(let ((weights '(1 2 3 5))
      (values  '(1 6 10 16))
      (capacity 7))
  (display "Knapsack Problem 2:") (newline)
  (display "  Capacity: ") (display capacity) (newline)
  (display "  Max value: ") (display (knapsack weights values capacity)) (newline))

; Example 3: all items fit
(let ((weights '(1 1 1))
      (values  '(10 20 30))
      (capacity 5))
  (display "Knapsack Problem 3 (all fit):") (newline)
  (display "  Max value: ") (display (knapsack weights values capacity)) (newline))
