; 75: Trie (Prefix Tree) using Association Lists
; A trie stores strings character-by-character for efficient prefix lookup.
; Each node is (children . is-end) where children is an alist.

(define (make-trie) (cons '() #f))
(define (trie-children t) (car t))
(define (trie-end? t) (cdr t))
(define (trie-set-end! t) (set-cdr! t #t))

(define (char-list str)
  ; Convert a symbol to a list of characters
  (let ((s (symbol->string str)))
    (let loop ((i 0))
      (if (= i (string-length s))
          '()
          (cons (string-ref s i) (loop (+ i 1)))))))

(define (trie-insert! trie chars)
  (if (null? chars)
      (trie-set-end! trie)
      (let ((child (assoc (car chars) (trie-children trie))))
        (if child
            (trie-insert! (cdr child) (cdr chars))
            (let ((new-node (make-trie)))
              (set-car! trie (cons (cons (car chars) new-node)
                                  (trie-children trie)))
              (trie-insert! new-node (cdr chars)))))))

(define (trie-search trie chars)
  (if (null? chars)
      (trie-end? trie)
      (let ((child (assoc (car chars) (trie-children trie))))
        (if child
            (trie-search (cdr child) (cdr chars))
            #f))))

(define (trie-starts-with? trie chars)
  (if (null? chars)
      #t
      (let ((child (assoc (car chars) (trie-children trie))))
        (if child
            (trie-starts-with? (cdr child) (cdr chars))
            #f))))

(define t (make-trie))
(trie-insert! t (char-list 'apple))
(trie-insert! t (char-list 'app))
(trie-insert! t (char-list 'banana))
(trie-insert! t (char-list 'band))

(display "Trie operations:") (newline)
(display "  search 'apple': ") (display (trie-search t (char-list 'apple))) (newline)
(display "  search 'app': ") (display (trie-search t (char-list 'app))) (newline)
(display "  search 'ap': ") (display (trie-search t (char-list 'ap))) (newline)
(display "  search 'banana': ") (display (trie-search t (char-list 'banana))) (newline)
(display "  prefix 'ban': ") (display (trie-starts-with? t (char-list 'ban))) (newline)
(display "  prefix 'cat': ") (display (trie-starts-with? t (char-list 'cat))) (newline)
