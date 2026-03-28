; 63: Practical example 3
; Range function
(define (range start end)
  (let loop ((i start) (acc (list)))
    (if (>= i end) (reverse acc) (loop (+ i 1) (cons i acc)))))
(display (range 0 5))  ; => (0 1 2 3 4)
(newline)
