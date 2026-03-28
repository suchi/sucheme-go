; 33: Exception handling
(display
  (guard (exn
    ((string? exn) (string-append "Error: " exn))
    (else "Unknown error"))
    (error "something went wrong")))
(newline)

; Nested guard
(display
  (guard (exn
    (else "outer"))
    (guard (inner-exn
      ((number? inner-exn) (+ inner-exn 1)))
      (raise 41))))
(newline)  ; => 42
