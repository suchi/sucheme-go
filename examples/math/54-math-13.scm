; 54: Euler's Totient Function (phi)
; phi(n) = count of integers in [1,n] that are coprime to n.
; Uses the product formula: phi(n) = n * product((1 - 1/p)) for each prime p dividing n.

(define (gcd a b)
  (if (= b 0) a (gcd b (modulo a b))))

; Brute-force totient by counting coprimes
(define (phi-brute n)
  (let loop ((i 1) (count 0))
    (if (> i n)
        count
        (loop (+ i 1) (if (= 1 (gcd i n)) (+ count 1) count)))))

; Efficient totient using prime factorization
(define (phi n)
  (let loop ((n n) (result n) (p 2))
    (cond ((> (* p p) n)
           (if (> n 1)
               (- result (/ result n))  ; n is a remaining prime factor
               result))
          ((= 0 (modulo n p))
           (let strip ((n n))
             (if (= 0 (modulo n p))
                 (strip (/ n p))
                 (loop n (- result (/ result p)) (+ p 1)))))
          (else (loop n result (+ p 1))))))

(display "Euler's totient phi(n) for n=1..20:") (newline)
(let loop ((n 1))
  (when (<= n 20)
    (display "  phi(") (display n) (display ") = ") (display (phi n))
    (display " [verify: ") (display (phi-brute n)) (display "]")
    (newline)
    (loop (+ n 1))))
