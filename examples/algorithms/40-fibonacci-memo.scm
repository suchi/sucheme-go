; 40: Memoized fibonacci
(define memo-table '())
(define (memo-fib n)
  (let ((cached (assoc n memo-table)))
    (if cached (cdr cached)
        (let ((result (if (< n 2) n
                         (+ (memo-fib (- n 1)) (memo-fib (- n 2))))))
          (set! memo-table (cons (cons n result) memo-table))
          result))))

(display (memo-fib 30))  ; => 832040
(newline)
