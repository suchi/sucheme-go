; 38: Quicksort
(define (quicksort lst)
  (if (or (null? lst) (null? (cdr lst))) lst
      (let ((pivot (car lst))
            (rest (cdr lst)))
        (let ((less (filter (lambda (x) (< x pivot)) rest))
              (greater (filter (lambda (x) (>= x pivot)) rest)))
          (append (quicksort less) (list pivot) (quicksort greater))))))

(display (quicksort '(5 2 8 1 9 3 7 4 6)))
(newline)  ; => (1 2 3 4 5 6 7 8 9)
