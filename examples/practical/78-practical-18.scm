; 78: Pattern Matching via Functions
; Implements a simple structural pattern matcher for lists.
; Pattern language:
;   _          -> wildcard (matches anything)
;   (? name)   -> capture variable (binds value to name)
;   (pat ...)  -> matches a list structurally
;   literal    -> matches by equal?

(define (capture? pat)
  (and (pair? pat) (equal? (car pat) '?) (pair? (cdr pat))))

(define (capture-name pat) (car (cdr pat)))

(define (match-pattern pat value bindings)
  (cond ((not bindings) #f)
        ((equal? pat '_) bindings)  ; wildcard
        ((capture? pat)
         (let ((name (capture-name pat)))
           (let ((existing (assoc name bindings)))
             (if existing
                 (if (equal? (cdr existing) value)
                     bindings
                     #f)
                 (cons (cons name value) bindings)))))
        ((and (pair? pat) (pair? value))
         (let ((head-result (match-pattern (car pat) (car value) bindings)))
           (match-pattern (cdr pat) (cdr value) head-result)))
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
; Patterns are quoted data: literal symbols match exactly, (? name) captures
(define (describe-shape shape)
  (pmatch shape
    ('(circle (? r))
     (lambda (b)
       (display "  Circle with radius ") (display (lookup 'r b)) (newline)))
    ('(rect (? w) (? h))
     (lambda (b)
       (display "  Rectangle ") (display (lookup 'w b))
       (display " x ") (display (lookup 'h b)) (newline)))
    ('(point (? x) (? y))
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
       ('(add (? a) (? b))
        (lambda (env)
          (display "  ") (display expr) (display " => ")
          (display (+ (lookup 'a env) (lookup 'b env))) (newline)))
       ('(mul (? a) (? b))
        (lambda (env)
          (display "  ") (display expr) (display " => ")
          (display (* (lookup 'a env) (lookup 'b env))) (newline)))
       ('(neg (? a))
        (lambda (env)
          (display "  ") (display expr) (display " => ")
          (display (- (lookup 'a env))) (newline)))))
   exprs))
