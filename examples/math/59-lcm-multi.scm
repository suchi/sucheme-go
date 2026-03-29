; 59: Least Common Multiple and GCD for Multiple Numbers
; LCM using the relation lcm(a,b) = |a*b| / gcd(a,b).
; Extends to multiple arguments via fold.

(define (gcd a b)
  (let ((a (abs a)) (b (abs b)))
    (if (= b 0) a (gcd b (modulo a b)))))

(define (abs x) (if (< x 0) (- x) x))

(define (lcm a b)
  (if (or (= a 0) (= b 0))
      0
      (/ (abs (* a b)) (gcd a b))))

(define (gcd-list lst)
  (let loop ((lst (cdr lst)) (acc (car lst)))
    (if (null? lst) acc (loop (cdr lst) (gcd acc (car lst))))))

(define (lcm-list lst)
  (let loop ((lst (cdr lst)) (acc (car lst)))
    (if (null? lst) acc (loop (cdr lst) (lcm acc (car lst))))))

(display "gcd(12, 18) = ") (display (gcd 12 18)) (newline)
(display "lcm(12, 18) = ") (display (lcm 12 18)) (newline)

(display "gcd(24, 36, 48) = ") (display (gcd-list '(24 36 48))) (newline)
(display "lcm(4, 6, 10) = ") (display (lcm-list '(4 6 10))) (newline)

(display "lcm(1..10) = ") (display (lcm-list '(1 2 3 4 5 6 7 8 9 10))) (newline)

; Property: gcd(a,b) * lcm(a,b) = |a*b|
(let ((a 15) (b 20))
  (display "gcd(15,20)*lcm(15,20) = ")
  (display (* (gcd a b) (lcm a b)))
  (display ", 15*20 = ") (display (* a b)) (newline))
