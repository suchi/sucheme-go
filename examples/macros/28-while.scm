; 28: While loop macro
(define-syntax while
  (syntax-rules ()
    ((while test body ...)
     (let loop ()
       (when test body ... (loop))))))

(define i 0)
(while (< i 5)
  (display i)
  (display " ")
  (set! i (+ i 1)))
(newline)
; => 0 1 2 3 4
