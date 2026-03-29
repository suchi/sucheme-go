; 77: Event System (Observer Pattern)
; A simple publish-subscribe event system using closures and mutation.

(define (make-event-bus)
  (let ((listeners '()))
    (define (on event-name handler)
      (set! listeners (cons (cons event-name handler) listeners)))
    (define (emit event-name . data)
      (for-each
       (lambda (pair)
         (when (equal? (car pair) event-name)
           (apply (cdr pair) data)))
       listeners))
    (define (listener-count)
      (length listeners))
    (lambda (msg . args)
      (cond ((equal? msg 'on) (apply on args))
            ((equal? msg 'emit) (apply emit args))
            ((equal? msg 'count) (listener-count))
            (else (display "Unknown message"))))))

(define bus (make-event-bus))

; Register listeners
(bus 'on 'greet
     (lambda (name)
       (display "  Hello, ") (display name) (display "!") (newline)))

(bus 'on 'greet
     (lambda (name)
       (display "  Welcome aboard, ") (display name) (display ".") (newline)))

(bus 'on 'calculate
     (lambda (a b)
       (display "  Sum: ") (display (+ a b)) (newline)))

(bus 'on 'calculate
     (lambda (a b)
       (display "  Product: ") (display (* a b)) (newline)))

(display "Registered ") (display (bus 'count)) (display " listeners") (newline)
(newline)

(display "Emitting 'greet with 'Alice:") (newline)
(bus 'emit 'greet 'Alice)

(newline)
(display "Emitting 'greet with 'Bob:") (newline)
(bus 'emit 'greet 'Bob)

(newline)
(display "Emitting 'calculate with 6 7:") (newline)
(bus 'emit 'calculate 6 7)
