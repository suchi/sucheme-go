; 29: Assert macro
(define-syntax assert
  (syntax-rules ()
    ((assert expr)
     (unless expr
       (error "assertion failed")))))

(assert (= (+ 1 1) 2))
(assert (string? "hello"))
(display "all assertions passed")
(newline)
