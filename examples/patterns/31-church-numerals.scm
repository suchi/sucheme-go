; 31: Church numerals
(define zero (lambda (f) (lambda (x) x)))
(define one (lambda (f) (lambda (x) (f x))))
(define two (lambda (f) (lambda (x) (f (f x)))))

(define (church-add m n)
  (lambda (f) (lambda (x) ((m f) ((n f) x)))))

(define (church->int n) ((n (lambda (x) (+ x 1))) 0))

(display (church->int zero))  ; => 0
(newline)
(display (church->int (church-add two two)))  ; => 4
(newline)
