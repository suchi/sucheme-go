; 60: Continued Fractions
; Represent rational numbers as continued fractions and convert back.
; A continued fraction [a0; a1, a2, ...] = a0 + 1/(a1 + 1/(a2 + ...))

(define (rational->cf numer denom)
  (if (= denom 0)
      '()
      (let ((q (quotient numer denom))
            (r (modulo numer denom)))
        (cons q (rational->cf denom r)))))

; Evaluate a continued fraction list back to a rational (as a pair)
(define (cf->rational cf)
  (if (null? cf)
      (cons 1 0)  ; infinity
      (if (null? (cdr cf))
          (cons (car cf) 1)
          (let ((rest (cf->rational (cdr cf))))
            (let ((p (car rest))
                  (q (cdr rest)))
              ;; a0 + q/p = (a0*p + q) / p
              (cons (+ (* (car cf) p) q) p))))))

(display "355/113 as CF: ")
(display (rational->cf 355 113))
(newline)

(display "CF [3 7 15 1] back to rational: ")
(display (cf->rational '(3 7 15 1)))
(newline)

(display "415/93 as CF: ")
(display (rational->cf 415 93))
(newline)

(display "CF [4 2 6 7] back to rational: ")
(display (cf->rational '(4 2 6 7)))
(newline)

; Golden ratio approximation: [1;1,1,1,1,1,1,1,1,1]
(display "Golden ratio CF [1 1 1 1 1 1 1 1 1 1]: ")
(display (cf->rational '(1 1 1 1 1 1 1 1 1 1)))
(newline)
