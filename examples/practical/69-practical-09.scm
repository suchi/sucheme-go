; 69: Practical example 9
; Group list into chunks
(define (chunk n lst)
  (if (null? lst) (list)
      (let loop ((l lst) (count 0) (current (list)))
        (cond ((null? l) (list (reverse current)))
              ((= count n) (cons (reverse current) (chunk n l)))
              (else (loop (cdr l) (+ count 1) (cons (car l) current)))))))
(display (chunk 2 (list 1 2 3 4 5)))
(newline)
