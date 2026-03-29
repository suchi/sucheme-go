; 56: Numerical Integration - Trapezoidal Rule and Simpson's Rule
; Approximate definite integrals using numerical methods.

(define (trapezoid f a b n)
  (let ((h (/ (- b a) n)))
    (let loop ((i 1) (sum (+ (f a) (f b))))
      (if (= i n)
          (* (/ h 2) sum)
          (loop (+ i 1) (+ sum (* 2 (f (+ a (* i h))))))))))

(define (simpson f a b n)
  ; n must be even
  (let ((h (/ (- b a) n)))
    (let loop ((i 1) (sum (+ (f a) (f b))))
      (if (= i n)
          (* (/ h 3) sum)
          (let ((coeff (if (= 0 (modulo i 2)) 2 4)))
            (loop (+ i 1) (+ sum (* coeff (f (+ a (* i h)))))))))))

; f(x) = x^2, integral from 0 to 1 = 1/3
(define (square x) (* x x))
(display "Integral of x^2 from 0 to 1:") (newline)
(display "  Trapezoid (n=10):  ") (display (trapezoid square 0 1 10)) (newline)
(display "  Trapezoid (n=100): ") (display (trapezoid square 0 1 100)) (newline)
(display "  Simpson (n=10):    ") (display (simpson square 0 1 10)) (newline)
(display "  Exact:             1/3") (newline)

; f(x) = x^3, integral from 0 to 2 = 4
(define (cube x) (* x x x))
(display "Integral of x^3 from 0 to 2:") (newline)
(display "  Trapezoid (n=100): ") (display (trapezoid cube 0 2 100)) (newline)
(display "  Simpson (n=10):    ") (display (simpson cube 0 2 10)) (newline)
(display "  Exact:             4") (newline)
