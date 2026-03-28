; 19: Greatest Common Divisor (Euclidean algorithm)
(define (gcd a b)
  (if (= b 0) a (gcd b (modulo a b))))

(display (gcd 48 18))   ; => 6
(newline)
(display (gcd 100 75))  ; => 25
(newline)
