; 27: Swap macro
(define-syntax swap!
  (syntax-rules ()
    ((swap! a b)
     (let ((tmp a)) (set! a b) (set! b tmp)))))

(define x 1)
(define y 2)
(swap! x y)
(display (list x y))  ; => (2 1)
(newline)
