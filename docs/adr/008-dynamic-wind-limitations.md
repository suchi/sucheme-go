# ADR-008: `dynamic-wind` の re-entrant 未対応

| 項目 | 値 |
|------|-----|
| 決定日 | 2026-03-28 |
| ステータス | 承認済み（有効・制限付き） |
| 関連ADR | [ADR-007](007-escape-continuation-only.md)（escape continuation の制約による） |
| 関連PR | [PR #9](https://github.com/suchi/sucheme-go/pull/9) |

## コンテキスト

R7RS の `dynamic-wind` は3つのサンク（before/body/after）を受け取り、継続の出入りに応じてbefore/afterを自動的に再実行することを要求する。具体的には:

- body に入る際に before が実行される
- body から出る際に after が実行される
- **継続で body に再入する際にも before が再実行される**
- **継続で body から再脱出する際にも after が再実行される**

ADR-007 で `call/cc` を escape continuation のみに限定したため、「再入」が発生しない。

## 決定

`dynamic-wind` は **escape continuation に対する before/after の実行のみをサポートする**。再入時の自動再実行は未対応とする。

## 理由

1. **ADR-007 の帰結**: escape continuation しかないため、body への再入が物理的に起こらない
2. **実用上十分**: 非ローカル脱出時のリソースクリーンアップ（ファイルクローズ、ロック解放等）はescape時のafterで対応可能
3. **正しい部分実装**: できないことを無理に実装するより、サポート範囲を明確にして正しく動作させる

## 実装（擬似コード）

実際の実装は `pkg/evaluator/continuation.go` を参照。以下は概要を示す擬似コード:

```go
type WindEntry struct {
    Before *types.Value  // before thunk（手続き）
    After  *types.Value  // after thunk（手続き）
}

// グローバルな windStack（LIFOスタック）で管理
// body 実行時: push → defer pop + after
// escape continuation 発火時: panic → recover で after が実行される
```

## 制約

- `(dynamic-wind before body after)` で取得した継続を再度呼び出しても、`before` は再実行されない
- wind スタックはグローバル変数で管理しており、シングルスレッド前提（non-thread-safe）
- R7RS準拠表では `dynamic-wind` は「部分対応」として記録

## 将来の変更

ADR-007 が更新されて full continuation が実装された場合、本ADRも連動して更新が必要。wind スタックを基に「現在の dynamic extent」と「ターゲットの dynamic extent」の差分を計算し、before/after を正しい順序で再実行する機構を追加する。
