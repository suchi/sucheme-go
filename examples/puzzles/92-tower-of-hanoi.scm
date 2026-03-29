; 92: Tower of Hanoi
; Classic recursive puzzle: move n disks from source to target via auxiliary peg.
; Demonstrates elegant recursion with 2^n - 1 moves.

(define *move-count* 0)

(define (hanoi n from to via)
  (when (> n 0)
    (hanoi (- n 1) from via to)
    (set! *move-count* (+ *move-count* 1))
    (display "  Move disk ") (display n)
    (display " from ") (display from)
    (display " to ") (display to)
    (newline)
    (hanoi (- n 1) via to from)))

(display "Tower of Hanoi with 3 disks:") (newline)
(set! *move-count* 0)
(hanoi 3 'A 'C 'B)
(display "  Total moves: ") (display *move-count*) (newline)

(newline)
(display "Tower of Hanoi with 4 disks:") (newline)
(set! *move-count* 0)
(hanoi 4 'A 'C 'B)
(display "  Total moves: ") (display *move-count*) (newline)

; Verify: moves for n disks = 2^n - 1
(newline)
(display "Move counts for n=1..8:") (newline)
(let loop ((n 1))
  (when (<= n 8)
    (set! *move-count* 0)
    (hanoi n 'A 'C 'B)
    (display "  n=") (display n) (display ": ")
    (display *move-count*) (display " moves")
    (display " (2^n-1 = ") (display (- (let pow ((k n) (r 1)) (if (= k 0) r (pow (- k 1) (* r 2)))) 1)) (display ")")
    (newline)
    (loop (+ n 1))))
