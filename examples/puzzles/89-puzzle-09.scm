; 89: Puzzle 9
; Caesar cipher
(define (caesar-encrypt s shift)
  (list->string
    (map (lambda (c)
           (if (char-alphabetic? c)
               (let* ((base (if (char-upper-case? c) 65 97))
                      (code (char->integer c)))
                 (integer->char (+ base (modulo (+ (- code base) shift) 26))))
               c))
         (string->list s))))
(display (caesar-encrypt "Hello World" 3))
(newline)
