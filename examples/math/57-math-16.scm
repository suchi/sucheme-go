; 57: Bisection Method for Root Finding
; Find roots of f(x)=0 in an interval [a,b] where f(a) and f(b) have opposite signs.

(define (bisect f a b tolerance max-iter)
  (let loop ((a a) (b b) (iter 0))
    (let ((mid (/ (+ a b) 2)))
      (cond ((> iter max-iter)
             (list 'max-iterations-reached mid))
            ((< (abs (- b a)) tolerance)
             (list 'converged mid iter))
            ((= 0 (f mid))
             (list 'exact mid iter))
            ((< (* (f a) (f mid)) 0)
             (loop a mid (+ iter 1)))
            (else
             (loop mid b (+ iter 1)))))))

(define (abs x) (if (< x 0) (- x) x))

; Root of x^2 - 2 = 0  =>  sqrt(2)
(display "sqrt(2) by bisection:") (newline)
(display (bisect (lambda (x) (- (* x x) 2)) 1 2 1/10000 100))
(newline)

; Root of x^3 - x - 1 = 0 (plastic number, approx 1.3247)
(display "Root of x^3 - x - 1:") (newline)
(display (bisect (lambda (x) (- (- (* x x x) x) 1)) 1 2 1/10000 100))
(newline)

; Root of x^2 - 3x + 2 = 0 => x=1 or x=2, find x near 1
(display "Root of x^2 - 3x + 2 near 0.5:") (newline)
(display (bisect (lambda (x) (+ (- (* x x) (* 3 x)) 2)) 0 3/2 1/10000 100))
(newline)
