; 83: Puzzle 3
; ROT13
(define (rot13-char c)
  (let ((code (char->integer c)))
    (cond
      ((and (>= code 65) (<= code 90))  ; A-Z
       (integer->char (+ 65 (modulo (+ (- code 65) 13) 26))))
      ((and (>= code 97) (<= code 122)) ; a-z
       (integer->char (+ 97 (modulo (+ (- code 97) 13) 26))))
      (else c))))
(define (rot13 s)
  (list->string (map rot13-char (string->list s))))
(display (rot13 "Hello"))  ; => Uryyb
(newline)
(display (rot13 (rot13 "Hello")))  ; => Hello
(newline)
