; 76: Simple Test Framework
; A lightweight testing framework using closures and continuations.

(define *test-pass* 0)
(define *test-fail* 0)
(define *test-errors* '())

(define (assert-equal test-name expected actual)
  (if (equal? expected actual)
      (begin
        (set! *test-pass* (+ *test-pass* 1))
        (display "  PASS: ") (display test-name) (newline))
      (begin
        (set! *test-fail* (+ *test-fail* 1))
        (set! *test-errors* (cons test-name *test-errors*))
        (display "  FAIL: ") (display test-name)
        (display " (expected ") (display expected)
        (display ", got ") (display actual)
        (display ")") (newline))))

(define (assert-true test-name value)
  (assert-equal test-name #t value))

(define (assert-false test-name value)
  (assert-equal test-name #f value))

(define (test-summary)
  (newline)
  (display "Results: ")
  (display *test-pass*) (display " passed, ")
  (display *test-fail*) (display " failed")
  (newline)
  (when (> *test-fail* 0)
    (display "Failed tests:") (newline)
    (for-each (lambda (name) (display "  - ") (display name) (newline))
              (reverse *test-errors*))))

; --- Run tests ---
(display "Testing list operations:") (newline)
(assert-equal "length of empty list" 0 (length '()))
(assert-equal "length of (1 2 3)" 3 (length '(1 2 3)))
(assert-equal "car of (1 2 3)" 1 (car '(1 2 3)))
(assert-equal "cdr of (1 2 3)" '(2 3) (cdr '(1 2 3)))
(assert-equal "append" '(1 2 3 4) (append '(1 2) '(3 4)))
(assert-equal "reverse" '(3 2 1) (reverse '(1 2 3)))
(assert-equal "map" '(2 4 6) (map (lambda (x) (* 2 x)) '(1 2 3)))
(assert-true "member found" (pair? (member 3 '(1 2 3 4))))
(assert-false "member not found" (member 5 '(1 2 3 4)))

(display "Testing arithmetic:") (newline)
(assert-equal "addition" 10 (+ 3 7))
(assert-equal "modulo" 1 (modulo 7 3))
(assert-equal "exponent via multiplication" 1024 (let loop ((n 10) (r 1)) (if (= n 0) r (loop (- n 1) (* r 2)))))

(test-summary)
