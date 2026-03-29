; 49: Math example 8
; Perfect square check
(define (perfect-square? n)
  (let loop ((i 0))
    (cond ((= (* i i) n) #t) ((> (* i i) n) #f) (else (loop (+ i 1))))))
(display (perfect-square? 49))  ; => #t
(newline)
(display (perfect-square? 50))  ; => #f
(newline)
