; 20: Prime number check
(define (prime? n)
  (define (check d)
    (cond
      ((> (* d d) n) #t)
      ((= (modulo n d) 0) #f)
      (else (check (+ d 1)))))
  (if (< n 2) #f (check 2)))

(define (primes-up-to n)
  (let loop ((i 2) (acc '()))
    (if (> i n)
        (reverse acc)
        (if (prime? i)
            (loop (+ i 1) (cons i acc))
            (loop (+ i 1) acc)))))

(display (primes-up-to 30))
(newline)
; => (2 3 5 7 11 13 17 19 23 29)
