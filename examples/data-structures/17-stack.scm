; 17: Stack using closures
(define (make-stack)
  (let ((items '()))
    (define (push! x)
      (set! items (cons x items)))
    (define (pop!)
      (if (null? items)
          (error "stack underflow")
          (let ((top (car items)))
            (set! items (cdr items))
            top)))
    (define (peek)
      (if (null? items) (error "empty stack") (car items)))
    (define (size) (length items))
    (define (dispatch msg)
      (cond
        ((equal? msg 'push!) push!)
        ((equal? msg 'pop!) pop!)
        ((equal? msg 'peek) peek)
        ((equal? msg 'size) size)
        (else (error "unknown message"))))
    dispatch))

(define s (make-stack))
((s 'push!) 1)
((s 'push!) 2)
((s 'push!) 3)
(display ((s 'size)))   ; => 3
(newline)
(display ((s 'pop!)))   ; => 3
(newline)
(display ((s 'peek)))   ; => 2
(newline)
