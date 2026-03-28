; 82: Puzzle 2
; Palindrome check
(define (palindrome? lst)
  (equal? lst (reverse lst)))
(display (palindrome? (list 1 2 3 2 1)))  ; => #t
(newline)
(display (palindrome? (list 1 2 3)))      ; => #f
(newline)
