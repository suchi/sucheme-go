; 71: RPN (Reverse Polish Notation) Calculator
; Evaluate expressions like "3 4 + 2 *" using a stack-based approach.

(define (rpn-eval tokens)
  (define (op? t)
    (or (equal? t '+) (equal? t '-) (equal? t '*) (equal? t '/)))
  (define (apply-op op a b)
    (cond ((equal? op '+) (+ a b))
          ((equal? op '-) (- a b))
          ((equal? op '*) (* a b))
          ((equal? op '/) (/ a b))))
  (let loop ((tokens tokens) (stack '()))
    (cond ((null? tokens)
           (car stack))
          ((op? (car tokens))
           (let ((b (car stack))
                 (a (car (cdr stack)))
                 (rest (cdr (cdr stack))))
             (loop (cdr tokens)
                   (cons (apply-op (car tokens) a b) rest))))
          (else
           (loop (cdr tokens)
                 (cons (car tokens) stack))))))

(display "3 4 + => ")
(display (rpn-eval '(3 4 +)))
(newline)

(display "3 4 + 2 * => ")
(display (rpn-eval '(3 4 + 2 *)))
(newline)

(display "5 1 2 + 4 * + 3 - => ")
(display (rpn-eval '(5 1 2 + 4 * + 3 -)))
(newline)

(display "10 2 / 3 + => ")
(display (rpn-eval '(10 2 / 3 +)))
(newline)
