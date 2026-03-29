; 80: Church Encoding - Representing Data with Pure Functions
; Numbers, booleans, and pairs encoded entirely as lambda functions.

; Church booleans
(define church-true  (lambda (t f) t))
(define church-false (lambda (t f) f))
(define church-if    (lambda (b then else) (b then else)))
(define church-and   (lambda (a b) (a b church-false)))
(define church-or    (lambda (a b) (a church-true b)))
(define church-not   (lambda (a) (a church-false church-true)))

; Decode church boolean to Scheme boolean
(define (church-bool->bool cb) (cb #t #f))

; Church numerals
(define church-zero (lambda (f x) x))
(define church-one  (lambda (f x) (f x)))
(define church-two  (lambda (f x) (f (f x))))

(define (church-succ n) (lambda (f x) (f (n f x))))
(define (church-add a b) (lambda (f x) (a f (b f x))))
(define (church-mul a b) (lambda (f x) (a (lambda (y) (b f y)) x)))

; Decode church numeral to Scheme number
(define (church-num->num cn) (cn (lambda (x) (+ x 1)) 0))

; Church pairs
(define (church-pair a b) (lambda (sel) (sel a b)))
(define (church-fst p) (p (lambda (a b) a)))
(define (church-snd p) (p (lambda (a b) b)))

; Build church numeral from integer
(define (num->church n)
  (if (= n 0)
      church-zero
      (church-succ (num->church (- n 1)))))

(display "Church Booleans:") (newline)
(display "  true AND false = ") (display (church-bool->bool (church-and church-true church-false))) (newline)
(display "  true OR false = ") (display (church-bool->bool (church-or church-true church-false))) (newline)
(display "  NOT true = ") (display (church-bool->bool (church-not church-true))) (newline)

(display "Church Numerals:") (newline)
(let ((three (church-succ church-two))
      (five (church-add church-two (church-succ church-two))))
  (display "  2 + 3 = ") (display (church-num->num (church-add church-two three))) (newline)
  (display "  2 * 3 = ") (display (church-num->num (church-mul church-two three))) (newline)
  (display "  5 = ") (display (church-num->num five)) (newline))

(display "Church Pairs:") (newline)
(let ((p (church-pair 42 99)))
  (display "  fst (42, 99) = ") (display (church-fst p)) (newline)
  (display "  snd (42, 99) = ") (display (church-snd p)) (newline))

; Build and decode larger number
(display "  num->church 7 => ") (display (church-num->num (num->church 7))) (newline)
