; 55: Modular Arithmetic - Extended GCD and Modular Inverse
; Extended Euclidean algorithm finds x,y such that a*x + b*y = gcd(a,b).
; Modular inverse of a mod m exists iff gcd(a,m)=1.

; Returns (gcd x y) as a list where a*x + b*y = gcd
(define (extended-gcd a b)
  (if (= b 0)
      (list a 1 0)
      (let ((r (extended-gcd b (modulo a b))))
        (let ((g (car r))
              (x (car (cdr r)))
              (y (car (cdr (cdr r)))))
          (list g y (- x (* (quotient a b) y)))))))

(define (mod-inverse a m)
  (let ((r (extended-gcd a m)))
    (if (= 1 (car r))
        (modulo (car (cdr r)) m)
        #f)))  ; no inverse exists

; Modular exponentiation: base^exp mod m (exp must be a non-negative integer)
(define (mod-pow base exp m)
  (if (< exp 0)
      (error "mod-pow: exponent must be a non-negative integer")
      (let loop ((base (modulo base m)) (exp exp) (result 1))
        (cond ((= exp 0) result)
              ((= 1 (modulo exp 2))
               (loop (modulo (* base base) m) (quotient exp 2) (modulo (* result base) m)))
              (else
               (loop (modulo (* base base) m) (quotient exp 2) result))))))

(display "Extended GCD(35, 15): ") (display (extended-gcd 35 15)) (newline)
(display "Extended GCD(120, 23): ") (display (extended-gcd 120 23)) (newline)

(display "Modular inverse of 3 mod 11: ") (display (mod-inverse 3 11)) (newline)
(display "Modular inverse of 7 mod 26: ") (display (mod-inverse 7 26)) (newline)
(display "Modular inverse of 6 mod 9: ") (display (mod-inverse 6 9)) (newline)

(display "2^10 mod 1000: ") (display (mod-pow 2 10 1000)) (newline)
(display "3^100 mod 97: ") (display (mod-pow 3 100 97)) (newline)

; Verify: a * inverse(a) mod m = 1
(let ((a 7) (m 26))
  (let ((inv (mod-inverse a m)))
    (display "7 * ") (display inv) (display " mod 26 = ")
    (display (modulo (* a inv) m)) (newline)))
