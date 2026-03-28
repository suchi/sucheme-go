; 84: Puzzle 4
; Number to words (1-10)
(define (number-word n)
  (list-ref (list "zero" "one" "two" "three" "four"
                  "five" "six" "seven" "eight" "nine" "ten") n))
(display (number-word 7))  ; => seven
(newline)
