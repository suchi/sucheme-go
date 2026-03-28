; 32: Y combinator (applicative-order)
(define Y
  (lambda (f)
    ((lambda (x) (f (lambda (v) ((x x) v))))
     (lambda (x) (f (lambda (v) ((x x) v)))))))

(define factorial
  (Y (lambda (self)
       (lambda (n)
         (if (= n 0) 1 (* n (self (- n 1))))))))

(display (factorial 10))  ; => 3628800
(newline)
