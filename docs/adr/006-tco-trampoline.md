# ADR-006: 末尾呼び出し最適化（TCO）にトランポリン方式を採用

| 項目 | 値 |
|------|-----|
| 決定日 | 2026-03-28 |
| ステータス | 承認済み（有効） |
| 関連ADR | [ADR-002](002-evaluation-strategy.md)（ツリーウォーキング方式の制約下での判断） |
| 関連PR | [PR #2](https://github.com/suchi/sucheme-go/pull/2) |

## コンテキスト

R7RS は末尾位置での手続き呼び出しが定数スタック空間で動作することを要求する（proper tail recursion）。ツリーウォーキングインタプリタ（ADR-002）では、再帰的な `Eval()` 呼び出しがGoのコールスタックを消費するため、素朴な実装では深い再帰でスタックオーバーフローが発生する。

### 検討した選択肢

1. **トランポリン方式**: 末尾位置から「次に評価すべき式と環境」（thunk）を返し、呼び出し元のwhileループで再評価
2. **CPS変換**: 全ての評価をCPS（継続渡しスタイル）に変換し、末尾呼び出しをGoの関数呼び出しに直接マッピング
3. **goroutine活用**: 各評価を個別のgoroutineで実行し、チャネルで通信

## 決定

**トランポリン方式を採用する。**

## 理由

1. **実装のシンプルさ**: thunk構造体（式+環境）とwhileループだけで実現できる
2. **既存コードとの親和性**: ツリーウォーキングインタプリタの構造を大きく変えずに追加できる
3. **デバッグ容易性**: 各ステップで式と環境が明示的に見える
4. **CPS変換の複雑さ回避**: CPS変換は全関数のシグネチャを変更し、コード全体の大幅な書き換えが必要

## 実装

```go
// pkg/evaluator/evaluator.go

type thunk struct {
    expr *types.Value // 次に評価する式
    env  *env.Env     // その式の環境
}

type evalResult struct {
    val   *types.Value
    thunk *thunk // nilでなければ再評価が必要
}

func Eval(expr *types.Value, e *env.Env) (*types.Value, error) {
    for {
        result := evalTCO(expr, e) // 1ステップ評価
        if result.thunk == nil {
            return result.val, nil // 最終値
        }
        expr = result.thunk.expr   // 末尾呼び出し → ループ継続
        e = result.thunk.env
    }
}
```

末尾位置（`if`の分岐、`begin`の最後の式、lambda本体の最後の式等）では `resultThunk()` を返し、非末尾位置では `resultVal()` を返す。

## 影響

- `if`, `cond`, `case`, `begin`, `let`, `letrec`, `do`, `when`, `unless`, `guard`, `case-lambda` の末尾位置でTCOが有効
- `or`/`and` の最終式でのTCO適用は今後の拡張対象（[ロードマップ「8. TCO完全対応」](../enhancement-roadmap.md)）
- バイトコードVMへの将来的な移行時は、TCOの実装方式も変更される
