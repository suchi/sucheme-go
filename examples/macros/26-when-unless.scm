; 26: when/unless macros
(define-syntax my-when
  (syntax-rules ()
    ((my-when test body ...)
     (if test (begin body ...)))))

(define-syntax my-unless
  (syntax-rules ()
    ((my-unless test body ...)
     (if (not test) (begin body ...)))))

(my-when #t
  (display "when is true")
  (newline))

(my-unless #f
  (display "unless is false")
  (newline))
