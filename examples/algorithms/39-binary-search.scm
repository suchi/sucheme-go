; 39: Binary search on sorted vector
(define (binary-search v target)
  (let loop ((lo 0) (hi (- (vector-length v) 1)))
    (if (> lo hi) -1
        (let ((mid (quotient (+ lo hi) 2)))
          (let ((val (vector-ref v mid)))
            (cond
              ((= val target) mid)
              ((< val target) (loop (+ mid 1) hi))
              (else (loop lo (- mid 1)))))))))

(define v (vector 1 3 5 7 9 11 13))
(display (binary-search v 7))  ; => 3
(newline)
(display (binary-search v 4))  ; => -1
(newline)
