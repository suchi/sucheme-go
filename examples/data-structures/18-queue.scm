; 18: Queue using two lists
(define (make-queue) (cons '() '()))
(define (enqueue! q x)
  (set-cdr! q (cons x (cdr q))))
(define (dequeue! q)
  (when (null? (car q))
    (set-car! q (reverse (cdr q)))
    (set-cdr! q '()))
  (let ((front (car (car q))))
    (set-car! q (cdr (car q)))
    front))
(define (queue-empty? q)
  (and (null? (car q)) (null? (cdr q))))

(define q (make-queue))
(enqueue! q 1)
(enqueue! q 2)
(enqueue! q 3)
(display (dequeue! q))  ; => 1
(newline)
(display (dequeue! q))  ; => 2
(newline)
