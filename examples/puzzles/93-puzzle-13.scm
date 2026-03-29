; 93: Maze Solver using Depth-First Search
; Represent a maze as a grid, find a path from start to goal.
; 0 = open, 1 = wall

(define maze
  '((0 0 1 0 0)
    (1 0 1 0 1)
    (0 0 0 0 0)
    (0 1 1 1 0)
    (0 0 0 0 0)))

(define (maze-ref maze r c)
  (list-ref (list-ref maze r) c))

(define (maze-rows maze) (length maze))
(define (maze-cols maze) (length (car maze)))

(define (in-bounds? maze r c)
  (and (>= r 0) (< r (maze-rows maze))
       (>= c 0) (< c (maze-cols maze))))

(define (member-pos? pos visited)
  (let loop ((v visited))
    (cond ((null? v) #f)
          ((and (= (car (car v)) (car pos))
                (= (car (cdr (car v))) (car (cdr pos))))
           #t)
          (else (loop (cdr v))))))

(define (solve-maze maze start goal)
  (define (dfs pos visited path)
    (let ((r (car pos)) (c (car (cdr pos))))
      (cond ((not (in-bounds? maze r c)) #f)
            ((= 1 (maze-ref maze r c)) #f)
            ((member-pos? pos visited) #f)
            ((and (= r (car goal)) (= c (car (cdr goal))))
             (reverse (cons pos path)))
            (else
             (let ((new-visited (cons pos visited))
                   (new-path (cons pos path)))
               (or (dfs (list (- r 1) c) new-visited new-path)
                   (dfs (list (+ r 1) c) new-visited new-path)
                   (dfs (list r (- c 1)) new-visited new-path)
                   (dfs (list r (+ c 1)) new-visited new-path)))))))
  (dfs start '() '()))

(define start '(0 0))
(define goal '(4 4))

(display "Maze (0=open, 1=wall):") (newline)
(for-each (lambda (row)
            (display "  ") (display row) (newline))
          maze)

(let ((path (solve-maze maze start goal)))
  (if path
      (begin
        (display "Path from ") (display start)
        (display " to ") (display goal) (display ":") (newline)
        (for-each (lambda (pos)
                    (display "  -> ") (display pos) (newline))
                  path))
      (begin
        (display "No path found!") (newline))))
