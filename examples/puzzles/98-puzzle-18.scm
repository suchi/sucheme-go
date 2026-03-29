; 98: Balanced Parentheses Checker and Generator
; Check if a sequence of 'open/'close symbols is balanced, and generate all balanced sequences of length 2n.

; Check balance over a list of 'open/'close symbols using a counter (stack simulation)
(define (balanced? chars)
  (let loop ((cs chars) (depth 0))
    (cond ((< depth 0) #f)
          ((null? cs) (= depth 0))
          ((equal? (car cs) 'open)
           (loop (cdr cs) (+ depth 1)))
          ((equal? (car cs) 'close)
           (loop (cdr cs) (- depth 1)))
          (else (loop (cdr cs) depth)))))

; Generate all balanced parentheses of n pairs
(define (gen-parens n)
  (define (gen open close current results)
    (cond ((and (= open 0) (= close 0))
           (cons (reverse current) results))
          (else
           (let ((r results))
             (when (> open 0)
               (set! r (gen (- open 1) close (cons 'open current) r)))
             (when (> close open)
               (set! r (gen open (- close 1) (cons 'close current) r)))
             r))))
  (reverse (gen n n '() '())))

(define (parens->string p)
  (apply string-append
         (map (lambda (x) (if (equal? x 'open) "(" ")")) p)))

; Check examples
(display "Balance checking:") (newline)
(display "  (()()) balanced? ") (display (balanced? '(open open close open close close))) (newline)
(display "  (()))  balanced? ") (display (balanced? '(open open close close close))) (newline)
(display "  empty  balanced? ") (display (balanced? '())) (newline)
(display "  )(     balanced? ") (display (balanced? '(close open))) (newline)

; Generate all balanced parens for n=1,2,3
(newline)
(let loop ((n 1))
  (when (<= n 4)
    (let ((results (gen-parens n)))
      (display "Balanced parens with ") (display n) (display " pairs (")
      (display (length results)) (display " total):") (newline)
      (for-each (lambda (p)
                  (display "  ") (display (parens->string p)) (newline))
                results)
      (loop (+ n 1)))))
