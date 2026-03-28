; 06: Let expressions
(display (let ((x 1) (y 2)) (+ x y)))  ; => 3
(newline)

; let* - sequential binding
(display (let* ((x 1) (y (+ x 1))) y))  ; => 2
(newline)

; Named let - loop
(display
  (let loop ((i 0) (sum 0))
    (if (= i 10) sum (loop (+ i 1) (+ sum i)))))  ; => 45
(newline)
