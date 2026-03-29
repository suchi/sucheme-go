; 95: Permutation Generator
; Generate all permutations of a list using backtracking.
; Demonstrates higher-order functions and list manipulation.

(define (remove-first x lst)
  (cond ((null? lst) '())
        ((equal? (car lst) x) (cdr lst))
        (else (cons (car lst) (remove-first x (cdr lst))))))

(define (permutations lst)
  (if (null? lst)
      '(())
      (apply append
             (map (lambda (x)
                    (map (lambda (perm) (cons x perm))
                         (permutations (remove-first x lst))))
                  lst))))

; Count permutations without generating them (n!)
(define (factorial n)
  (let loop ((i 1) (acc 1))
    (if (> i n) acc (loop (+ i 1) (* acc i)))))

(display "Permutations of (1 2 3):") (newline)
(for-each (lambda (p) (display "  ") (display p) (newline))
          (permutations '(1 2 3)))

(display "Number of permutations of (a b c d): ")
(display (length (permutations '(a b c d))))
(display " (4! = ") (display (factorial 4)) (display ")")
(newline)

; Permutations of a string-like structure
(display "Permutations of (X Y Z):") (newline)
(for-each (lambda (p) (display "  ") (display p) (newline))
          (permutations '(X Y Z)))
