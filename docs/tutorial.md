# Sucheme チュートリアル

Sucheme で Scheme プログラミングを始めるためのステップバイステップガイドです。

## 目次

1. [はじめに](#はじめに)
2. [基本的な式と型](#基本的な式と型)
3. [変数と関数](#変数と関数)
4. [条件分岐](#条件分岐)
5. [リスト操作](#リスト操作)
6. [高階関数](#高階関数)
7. [クロージャ](#クロージャ)
8. [再帰と末尾呼び出し最適化](#再帰と末尾呼び出し最適化)
9. [マクロ](#マクロ)
10. [例外処理](#例外処理)
11. [パラメータと動的束縛](#パラメータと動的束縛)
12. [ライブラリシステム](#ライブラリシステム)
13. [実践例: アルゴリズム](#実践例-アルゴリズム)
14. [case-lambda（多重ディスパッチ）](#case-lambda多重ディスパッチ)
15. [Web REPL](#web-repl)
16. [次のステップ](#次のステップ)

---

## はじめに

### インストール

```bash
go install github.com/suchi/sucheme-go/cmd/sucheme@latest
```

### REPL で試す

```bash
$ sucheme
sucheme> (+ 1 2 3)
6
sucheme> (display "Hello, World!")
Hello, World!
```

### ファイルを実行する

```bash
$ sucheme examples/basics/01-hello.scm
Hello, World!
```

### 式を直接評価する

```bash
$ sucheme -e '(* 6 7)'
42
```

---

## 基本的な式と型

Scheme では、すべてがS式 (S-expression) です。`(演算子 引数 ...)` の形式で記述します。

### 数値

```scheme
sucheme> 42        ; 整数
42
sucheme> 3.14      ; 浮動小数点数
3.14
sucheme> (+ 1 2)   ; 加算
3
sucheme> (* 3 4)   ; 乗算
12
sucheme> (- 10 3)  ; 減算
7
sucheme> (/ 15 4)  ; 除算
3.75
```

### 文字列

```scheme
sucheme> "hello"
"hello"
sucheme> (string-length "hello")
5
sucheme> (string-append "hello" " " "world")
"hello world"
sucheme> (substring "hello" 1 3)
"el"
```

### 真偽値

```scheme
sucheme> #t         ; 真
#t
sucheme> #f         ; 偽
#f
sucheme> (= 1 1)
#t
sucheme> (< 3 5)
#t
sucheme> (not #f)
#t
```

### シンボルとクォート

```scheme
sucheme> 'hello         ; シンボル
hello
sucheme> '(1 2 3)       ; クォートされたリスト（評価されない）
(1 2 3)
sucheme> (quote (a b))  ; quote の明示的な使用
(a b)
```

---

## 変数と関数

### 変数定義

```scheme
(define x 10)
(define greeting "Hello")
(display x)        ; 出力: 10
(display greeting) ; 出力: Hello
```

### 関数定義

```scheme
;; 短縮形
(define (square x) (* x x))
(square 5) ; => 25

;; lambda を使った同等の定義
(define square (lambda (x) (* x x)))

;; 複数引数
(define (add a b) (+ a b))
(add 3 4) ; => 7

;; 無名関数の直接適用
((lambda (x) (* x 2)) 21) ; => 42
```

### 可変長引数

```scheme
;; ドット記法で残余引数を受け取る
(define (my-list . args) args)
(my-list 1 2 3) ; => (1 2 3)

;; 固定引数＋残余引数
(define (f a b . rest) (list a b rest))
(f 1 2 3 4 5) ; => (1 2 (3 4 5))
```

### let 式（ローカル変数）

```scheme
(let ((x 10)
      (y 20))
  (+ x y)) ; => 30

;; let* は前の束縛を参照できる
(let* ((x 1)
       (y (+ x 1)))
  y) ; => 2
```

---

## 条件分岐

### if

```scheme
(if (> 3 2)
    "yes"
    "no") ; => "yes"
```

### cond

```scheme
(define (classify n)
  (cond
    ((< n 0)  "negative")
    ((= n 0)  "zero")
    (else     "positive")))

(classify -5) ; => "negative"
(classify 0)  ; => "zero"
(classify 7)  ; => "positive"
```

### and / or

```scheme
(and 1 2 3)     ; => 3（最後の真値）
(and 1 #f 3)    ; => #f（短絡評価）
(or #f #f 42)   ; => 42（最初の真値）
(or #f #f #f)   ; => #f
```

### when / unless

```scheme
(when (> 5 3)
  (display "5 is greater")
  (newline))

(unless (< 5 3)
  (display "5 is not less")
  (newline))
```

---

## リスト操作

リストは Scheme の基本データ構造です。

### リストの構築

```scheme
(list 1 2 3)     ; => (1 2 3)
(cons 0 '(1 2))  ; => (0 1 2)  ← 先頭に追加
'(a b c)         ; => (a b c)  ← クォートで直接記述
```

### リストの分解

```scheme
(car '(a b c))   ; => a       ← 先頭要素
(cdr '(a b c))   ; => (b c)   ← 先頭以外
(cadr '(a b c))  ; => b       ← 2番目
```

### 基本操作

```scheme
(length '(1 2 3 4 5))     ; => 5
(append '(1 2) '(3 4))    ; => (1 2 3 4)
(reverse '(1 2 3))        ; => (3 2 1)
(list-ref '(a b c d) 2)   ; => c
(null? '())               ; => #t
(pair? '(1 2))            ; => #t
```

---

## 高階関数

関数を引数に取る関数、あるいは関数を返す関数を高階関数と呼びます。

### map

リストの各要素に関数を適用した新しいリストを返します。

```scheme
(map (lambda (x) (* x x)) '(1 2 3 4 5))
; => (1 4 9 16 25)

;; 注: R7RS では (map f lst1 lst2 ...) のように複数リストを渡す map もありますが、
;; Sucheme の現在の実装では 1 リスト版のみ対応です。
;; 複数リストを渡してもエラーにはなりませんが、2つ目以降は無視されるため、
;; 必ず (map f lst) の形で使用してください。
```

### filter

条件を満たす要素だけを残します。

```scheme
(filter (lambda (x) (> x 3)) '(1 2 3 4 5))
; => (4 5)

(filter even? '(1 2 3 4 5 6))
; => (2 4 6)
```

### for-each

副作用のための反復処理です（返り値なし）。

```scheme
(for-each
  (lambda (x) (display x) (display " "))
  '(a b c))
; 出力: a b c
```

### fold (畳み込み)

```scheme
;; foldl: 左畳み込み
(define (foldl f init lst)
  (if (null? lst) init
      (foldl f (f init (car lst)) (cdr lst))))

(foldl + 0 '(1 2 3 4 5)) ; => 15
```

### 関数合成

```scheme
(define (compose f g)
  (lambda (x) (f (g x))))

(define add1-then-double (compose (lambda (x) (* x 2)) (lambda (x) (+ x 1))))
(add1-then-double 3) ; => 8  (3+1=4, 4*2=8)
```

---

## クロージャ

クロージャは、定義された環境を捕捉する関数です。状態を持つオブジェクトを作れます。

### カウンター

```scheme
(define (make-counter)
  (let ((n 0))
    (lambda ()
      (set! n (+ n 1))
      n)))

(define c (make-counter))
(c) ; => 1
(c) ; => 2
(c) ; => 3
```

### アキュムレーター

```scheme
(define (make-accumulator initial)
  (let ((total initial))
    (lambda (amount)
      (set! total (+ total amount))
      total)))

(define acc (make-accumulator 100))
(acc 25)  ; => 125
(acc -10) ; => 115
```

---

## 再帰と末尾呼び出し最適化

Sucheme はトランポリン方式の末尾呼び出し最適化 (TCO) を実装しています。
末尾位置の再帰呼び出しはスタックを消費せず、何百万回でも実行できます。

### 通常の再帰（非末尾再帰）

```scheme
(define (factorial n)
  (if (= n 0) 1
      (* n (factorial (- n 1)))))

(factorial 10) ; => 3628800
```

### 末尾再帰（TCO 対応）

```scheme
;; アキュムレータパターン
(define (factorial-iter n)
  (define (loop i acc)
    (if (> i n) acc
        (loop (+ i 1) (* acc i))))
  (loop 1 1))

(factorial-iter 10) ; => 3628800
```

### 名前付き let を使った末尾再帰

```scheme
(define (sum-to n)
  (let loop ((i 0) (acc 0))
    (if (> i n) acc
        (loop (+ i 1) (+ acc i)))))

(sum-to 100000) ; => 5000050000（スタックオーバーフローなし）
```

### do ループ

```scheme
(define (factorial-do n)
  (do ((i 1 (+ i 1))
       (result 1 (* result i)))
      ((> i n) result)))

(factorial-do 10) ; => 3628800
```

---

## マクロ

`syntax-rules` による衛生的マクロ (hygienic macros) を使って、
構文を自分で拡張できます。

### 基本的なマクロ

```scheme
(define-syntax my-when
  (syntax-rules ()
    ((my-when test body ...)
     (if test (begin body ...)))))

(my-when #t
  (display "hello")
  (newline))
```

### swap マクロ

`set!` を使うため、変数（識別子）同士の入れ替え専用です。

```scheme
(define-syntax swap!
  (syntax-rules ()
    ((swap! a b)
     (let ((tmp a))
       (set! a b)
       (set! b tmp)))))

(define x 1)
(define y 2)
(swap! x y)
(display x) ; 出力: 2
(display y) ; 出力: 1
```

### while ループマクロ

```scheme
(define-syntax while
  (syntax-rules ()
    ((while test body ...)
     (let loop ()
       (when test body ... (loop))))))

(define n 5)
(while (> n 0)
  (display n)
  (display " ")
  (set! n (- n 1)))
; 出力: 5 4 3 2 1
```

---

## 例外処理

`guard` で例外をキャッチし、`raise` / `error` でスローします。

### 基本的な例外処理

```scheme
(guard (exn
    ((string? exn) (string-append "Error: " exn))
    (else "Unknown error"))
  (error "something went wrong"))
; => "Error: something went wrong"
```

### raise で値を投げる

```scheme
(guard (exn
    ((number? exn) (+ exn 1)))
  (raise 41))
; => 42
```

### with-exception-handler

`with-exception-handler` は引数を2つ（handler と thunk）だけ取ります。
継続可能な再開には `raise-continuable` を使います。

```scheme
(with-exception-handler
  (lambda (exn)
    (display "caught: ")
    (display exn)
    (newline)
    42)  ; handler の戻り値が raise-continuable の結果になる
  (lambda ()
    (raise-continuable "oops")))
; => "caught: oops" と表示され、式全体の値は 42
```

---

## パラメータと動的束縛

`make-parameter` と `parameterize` で動的束縛 (dynamic binding) を実現します。

```scheme
(define current-lang (make-parameter "en"))

(display (current-lang)) ; 出力: en

(parameterize ((current-lang "ja"))
  (display (current-lang))) ; 出力: ja

(display (current-lang)) ; 出力: en（元に戻る）
```

ネストも可能です:

```scheme
(define indent (make-parameter 0))

(parameterize ((indent 2))
  (display (indent)) ; 出力: 2
  (parameterize ((indent 4))
    (display (indent))) ; 出力: 4
  (display (indent))) ; 出力: 2
```

---

## ライブラリシステム

R7RS の `define-library` でモジュールを定義できます。

### ライブラリの定義

```scheme
(define-library (mylib math)
  (import (scheme base))
  (export square cube)
  (begin
    (define (square x) (* x x))
    (define (cube x) (* x x x))))
```

### ライブラリのインポート

```scheme
(import (mylib math))
(square 5) ; => 25
(cube 3)   ; => 27
```

### インポートの制御

```scheme
;; 特定のシンボルだけインポート
(import (only (mylib math) square))

;; 特定のシンボルを除外
(import (except (mylib math) cube))

;; プレフィックス付きインポート
(import (prefix (mylib math) m:))
(m:square 5) ; => 25

;; シンボル名の変更
(import (rename (mylib math) (square sq)))
(sq 5) ; => 25
```

---

## 実践例: アルゴリズム

### クイックソート

```scheme
(define (quicksort lst)
  (if (or (null? lst) (null? (cdr lst))) lst
      (let ((pivot (car lst))
            (rest (cdr lst)))
        (let ((less (filter (lambda (x) (< x pivot)) rest))
              (greater (filter (lambda (x) (>= x pivot)) rest)))
          (append (quicksort less)
                  (list pivot)
                  (quicksort greater))))))

(quicksort '(5 2 8 1 9 3 7 4 6))
; => (1 2 3 4 5 6 7 8 9)
```

### フィボナッチ数列（メモ化）

```scheme
(define memo-table '())

(define (memo-fib n)
  (let ((cached (assv n memo-table)))
    (if cached
        (cdr cached)
        (let ((result (if (< n 2) n
                          (+ (memo-fib (- n 1))
                             (memo-fib (- n 2))))))
          (set! memo-table (cons (cons n result) memo-table))
          result))))

(memo-fib 30) ; => 832040
```

### 素数判定

```scheme
(define (prime? n)
  (if (< n 2) #f
      (let loop ((i 2))
        (cond
          ((> (* i i) n) #t)
          ((= (modulo n i) 0) #f)
          (else (loop (+ i 1)))))))

(filter prime? '(2 3 4 5 6 7 8 9 10 11 12 13))
; => (2 3 5 7 11 13)
```

---

## case-lambda（多重ディスパッチ）

引数の数に応じて異なる処理を行う関数を定義できます。

```scheme
(define greet
  (case-lambda
    (() "Hello!")
    ((name) (string-append "Hello, " name "!"))
    ((name greeting) (string-append greeting ", " name "!"))))

(greet)               ; => "Hello!"
(greet "Alice")       ; => "Hello, Alice!"
(greet "Alice" "Hi")  ; => "Hi, Alice!"
```

---

## Web REPL

ブラウザ上でも Sucheme を試すことができます。WebAssembly でコンパイルされた
インタプリタがブラウザ内で動作します。

### ローカルでの起動

```bash
# WASM ビルド
GOOS=js GOARCH=wasm go build -o web-ui/sucheme.wasm ./cmd/wasm/
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web-ui/

# HTTP サーバーで配信
# 注: .wasm の MIME タイプ (application/wasm) が必要です。
# Python 3.9+ なら自動対応ですが、古いバージョンでは
# WASM のロードに失敗する場合があります。
cd web-ui && python3 -m http.server 8080
```

ブラウザで `http://localhost:8080` を開くと Web REPL が使えます。

---

## 次のステップ

- `examples/` ディレクトリに100以上のサンプルプログラムがあります
- `docs/r7rs-compliance.md` で R7RS 対応状況を確認できます
- `docs/onboarding.md` で開発者向けのガイドを参照できます
- [GitHub Actions](https://github.com/suchi/sucheme-go/actions) でCI/CDの状況を確認できます
