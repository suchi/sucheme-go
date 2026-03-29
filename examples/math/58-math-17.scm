; 58: Matrix Operations using Lists of Lists
; Represent matrices as row-major lists. Supports addition, scaling,
; multiplication, transpose, and determinant.

(define (matrix-ref m i j)
  (list-ref (list-ref m i) j))

(define (matrix-rows m) (length m))
(define (matrix-cols m) (length (car m)))

(define (matrix-transpose m)
  (let ((rows (matrix-rows m))
        (cols (matrix-cols m)))
    (let loop-c ((j 0))
      (if (= j cols)
          '()
          (cons (let loop-r ((i 0))
                  (if (= i rows)
                      '()
                      (cons (matrix-ref m i j) (loop-r (+ i 1)))))
                (loop-c (+ j 1)))))))

(define (matrix-add a b)
  (map (lambda (ra rb) (map + ra rb)) a b))

(define (matrix-scale k m)
  (map (lambda (row) (map (lambda (x) (* k x)) row)) m))

(define (dot-product a b)
  (apply + (map * a b)))

(define (matrix-multiply a b)
  (let ((bt (matrix-transpose b)))
    (map (lambda (row-a)
           (map (lambda (col-b)
                  (dot-product row-a col-b))
                bt))
         a)))

; 2x2 determinant
(define (det2 m)
  (- (* (matrix-ref m 0 0) (matrix-ref m 1 1))
     (* (matrix-ref m 0 1) (matrix-ref m 1 0))))

(define A '((1 2) (3 4)))
(define B '((5 6) (7 8)))

(display "A = ") (display A) (newline)
(display "B = ") (display B) (newline)
(display "A + B = ") (display (matrix-add A B)) (newline)
(display "2*A = ") (display (matrix-scale 2 A)) (newline)
(display "A * B = ") (display (matrix-multiply A B)) (newline)
(display "Transpose(A) = ") (display (matrix-transpose A)) (newline)
(display "det(A) = ") (display (det2 A)) (newline)

; 3x3 example
(define C '((1 0 0) (0 1 0) (0 0 1)))  ; identity
(define D '((1 2 3) (4 5 6) (7 8 9)))
(display "I * D = ") (display (matrix-multiply C D)) (newline)
