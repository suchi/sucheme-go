; 35: String operations
(define (string-contains? s sub)
  (let ((slen (string-length s))
        (sublen (string-length sub)))
    (let loop ((i 0))
      (cond
        ((> (+ i sublen) slen) #f)
        ((string=? (substring s i (+ i sublen)) sub) #t)
        (else (loop (+ i 1)))))))

(display (string-contains? "hello world" "world"))  ; => #t
(newline)
(display (string-contains? "hello world" "xyz"))    ; => #f
(newline)

; String split (simple)
(define (string-split s ch)
  (let ((len (string-length s)))
    (let loop ((i 0) (start 0) (acc '()))
      (cond
        ((= i len) (reverse (cons (substring s start len) acc)))
        ((char=? (string-ref s i) ch)
         (loop (+ i 1) (+ i 1) (cons (substring s start i) acc)))
        (else (loop (+ i 1) start acc))))))

(display (string-split "a,b,c,d" #\,))  ; => (a b c d)
(newline)
