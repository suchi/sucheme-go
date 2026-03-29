; 78: Pattern Matching via Macros / Functions
; Implements a simple structural pattern matcher for lists.
; Patterns: literal values, '_ as wildcard, symbols as captures.

(define (match-pattern pat value bindings)
  (cond ((equal? pat '_) bindings)  ; wildcard
        ((and (symbol? pat) (not (null? bindings)))
         ; Check if already bound
         (let ((existing (assoc pat bindings)))
           (if existing
               (if (equal? (cdr existing) value)
                   bindings
                   #f)
               (cons (cons pat value) bindings))))
        ((symbol? pat)
         (cons (cons pat value) bindings))
        ((and (pair? pat) (pair? value))
         (let ((head-result (match-pattern (car pat) (car value) bindings)))
           (if head-result
               (match-pattern (cdr pat) (cdr value) head-result)
               #f)))
        ((and (null? pat) (null? value)) bindings)
        ((equal? pat value) bindings)
        (else #f)))

(define (pmatch value . clauses)
  (let loop ((clauses clauses))
    (if (null? clauses)
        (begin (display "No match found") (newline) #f)
        (let* ((clause (car clauses))
               (pattern (car clause))
               (body (car (cdr clause)))
               (result (match-pattern pattern value '())))
          (if result
              (body result)
              (loop (cdr clauses)))))))

(define (lookup key bindings)
  (cdr (assoc key bindings)))

; Example: classify shapes
(define (describe-shape shape)
  (pmatch shape
    (('circle r)
     (lambda (b)
       (display "  Circle with radius ") (display (lookup 'r b)) (newline)))
    (('rect w h)
     (lambda (b)
       (display "  Rectangle ") (display (lookup 'w b))
       (display " x ") (display (lookup 'h b)) (newline)))
    (('point x y)
     (lambda (b)
       (display "  Point at (") (display (lookup 'x b))
       (display ", ") (display (lookup 'y b)) (display ")") (newline)))))

(display "Shape pattern matching:") (newline)
(describe-shape '(circle 5))
(describe-shape '(rect 3 4))
(describe-shape '(point 10 20))

; Example: simple expression evaluator via matching
(display "Expression matching:") (newline)
(let ((exprs (list '(add 3 4) '(mul 5 6) '(neg 7))))
  (for-each
   (lambda (expr)
     (pmatch expr
       (('add a b)
        (lambda (b-env)
          (display "  ") (display expr) (display " => ")
          (display (+ (lookup 'a b-env) (lookup 'b b-env))) (newline)))
       (('mul a b)
        (lambda (b-env)
          (display "  ") (display expr) (display " => ")
          (display (* (lookup 'a b-env) (lookup 'b b-env))) (newline)))
       (('neg a)
        (lambda (b-env)
          (display "  ") (display expr) (display " => ")
          (display (- (lookup 'a b-env))) (newline)))))
   exprs))
