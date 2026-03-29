; 70: Practical example 10
; Repeat value n times
(define (repeat n val)
  (let loop ((i 0) (acc (list)))
    (if (= i n) (reverse acc) (loop (+ i 1) (cons val acc)))))
(display (repeat 5 "x"))  ; => ("x" "x" "x" "x" "x")
(newline)
