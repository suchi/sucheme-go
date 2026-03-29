; 74: State Machine
; A finite state machine for validating simple patterns.
; This FSM recognizes binary strings that contain "101" as a substring.

(define (make-fsm transitions accept-states)
  (list transitions accept-states))

(define (fsm-transitions fsm) (car fsm))
(define (fsm-accept-states fsm) (car (cdr fsm)))

(define (fsm-step transitions state input)
  (let loop ((ts transitions))
    (if (null? ts)
        #f  ; no transition found
        (let ((t (car ts)))
          (if (and (equal? (car t) state)
                   (equal? (car (cdr t)) input))
              (car (cdr (cdr t)))
              (loop (cdr ts)))))))

(define (fsm-run fsm input-list)
  (let ((transitions (fsm-transitions fsm))
        (accept (fsm-accept-states fsm)))
    (let loop ((state 'q0) (input input-list))
      (cond ((null? input)
             (let check ((as accept))
               (cond ((null? as) #f)
                     ((equal? state (car as)) #t)
                     (else (check (cdr as))))))
            (else
             (let ((next (fsm-step transitions state (car input))))
               (if next
                   (loop next (cdr input))
                   #f)))))))

; FSM that recognizes binary strings containing "101"
(define contains-101
  (make-fsm
   '((q0 0 q0) (q0 1 q1)     ; start state
     (q1 0 q2) (q1 1 q1)     ; seen "1"
     (q2 0 q0) (q2 1 q3)     ; seen "10"
     (q3 0 q3) (q3 1 q3))    ; seen "101" - accept, absorb all
   '(q3)))

(define (test-fsm input)
  (display "  \"") (display input) (display "\" => ")
  (display (if (fsm-run contains-101 input) "ACCEPT" "REJECT"))
  (newline))

(display "FSM: does binary string contain '101'?") (newline)
(test-fsm '(1 0 1))
(test-fsm '(1 0 1 0))
(test-fsm '(0 0 1 0 1 1 0))
(test-fsm '(1 1 1))
(test-fsm '(0 0 0))
(test-fsm '(1 0 0))
(test-fsm '(1 1 0 1 1))
