; 25: Object-oriented style with closures
(define (make-point x y)
  (define (get-x) x)
  (define (get-y) y)
  (define (distance-to other)
    (let ((dx (- (other 'x) x))
          (dy (- (other 'y) y)))
      (sqrt-approx (+ (* dx dx) (* dy dy)))))
  (define (to-string)
    (string-append "(" (number->string x) ", " (number->string y) ")"))
  (define (dispatch msg)
    (cond
      ((equal? msg 'x) (get-x))
      ((equal? msg 'y) (get-y))
      ((equal? msg 'to-string) (to-string))
      (else (error "unknown message"))))
  dispatch)

; Simple sqrt approximation
(define (sqrt-approx x)
  (let loop ((guess 1.0) (i 0))
    (if (= i 20) guess
        (loop (/ (+ guess (/ x guess)) 2.0) (+ i 1)))))

(define p1 (make-point 3 4))
(display (p1 'x))          ; => 3
(newline)
(display (p1 'to-string))  ; => (3, 4)
(newline)
