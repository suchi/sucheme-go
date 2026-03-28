; 21: Fast exponentiation
(define (fast-expt base exp)
  (cond
    ((= exp 0) 1)
    ((= (modulo exp 2) 0)
     (let ((half (fast-expt base (/ exp 2))))
       (* half half)))
    (else (* base (fast-expt base (- exp 1))))))

(display (fast-expt 2 10))  ; => 1024
(newline)
(display (fast-expt 3 5))   ; => 243
(newline)
