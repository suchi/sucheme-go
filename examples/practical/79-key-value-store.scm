; 79: Simple Key-Value Store with Transactions
; Demonstrates closures for encapsulation and mutation for state management.
; Uses tombstone markers for transactional deletes.

; Unique sentinel object - compared by identity (eq?) not structure (equal?)
(define *tombstone* (list '*deleted*))

(define (make-kvstore)
  (let ((data '())
        (transaction #f)
        (txn-data '()))
    (define (remove-key key lst)
      (cond ((null? lst) '())
            ((equal? (car (car lst)) key) (remove-key key (cdr lst)))
            (else (cons (car lst) (remove-key key (cdr lst))))))
    (define (get key)
      (if transaction
          (let ((txn-entry (assoc key txn-data)))
            (if txn-entry
                (if (eq? (cdr txn-entry) *tombstone*)
                    #f  ; deleted in transaction
                    (cdr txn-entry))
                (let ((entry (assoc key data)))
                  (if entry (cdr entry) #f))))
          (let ((entry (assoc key data)))
            (if entry (cdr entry) #f))))
    (define (put! key value)
      (if transaction
          (set! txn-data (cons (cons key value) (remove-key key txn-data)))
          (set! data (cons (cons key value) (remove-key key data)))))
    (define (delete! key)
      (if transaction
          ; Store tombstone in txn-data to shadow base data
          (set! txn-data (cons (cons key *tombstone*) (remove-key key txn-data)))
          (set! data (remove-key key data))))
    (define (begin-txn!)
      (set! transaction #t)
      (set! txn-data '()))
    (define (commit!)
      (let ((saved txn-data))
        (set! transaction #f)
        (set! txn-data '())
        (for-each (lambda (pair)
                    (if (eq? (cdr pair) *tombstone*)
                        (delete! (car pair))
                        (put! (car pair) (cdr pair))))
                  saved))
      (display "  Transaction committed.") (newline))
    (define (rollback!)
      (set! transaction #f)
      (set! txn-data '())
      (display "  Transaction rolled back.") (newline))
    (lambda (msg . args)
      (cond ((equal? msg 'get) (apply get args))
            ((equal? msg 'put) (apply put! args))
            ((equal? msg 'delete) (apply delete! args))
            ((equal? msg 'begin) (begin-txn!))
            ((equal? msg 'commit) (commit!))
            ((equal? msg 'rollback) (rollback!))))))

(define db (make-kvstore))

(display "Key-Value Store Demo:") (newline)

(db 'put 'name "Alice")
(db 'put 'age 30)
(display "  name = ") (display (db 'get 'name)) (newline)
(display "  age = ") (display (db 'get 'age)) (newline)

; Transaction: commit
(display "Begin transaction (update age to 31)...") (newline)
(db 'begin)
(db 'put 'age 31)
(display "  age in txn = ") (display (db 'get 'age)) (newline)
(db 'commit)
(display "  age after commit = ") (display (db 'get 'age)) (newline)

; Transaction: rollback
(display "Begin transaction (update name to Bob)...") (newline)
(db 'begin)
(db 'put 'name "Bob")
(display "  name in txn = ") (display (db 'get 'name)) (newline)
(db 'rollback)
(display "  name after rollback = ") (display (db 'get 'name)) (newline)

; Transaction: delete within transaction
(display "Begin transaction (delete age)...") (newline)
(db 'begin)
(db 'delete 'age)
(display "  age in txn (deleted) = ") (display (db 'get 'age)) (newline)
(db 'commit)
(display "  age after commit = ") (display (db 'get 'age)) (newline)
