# ADR-007: `call/cc` を escape continuation のみで実装

| 項目 | 値 |
|------|-----|
| 決定日 | 2026-03-28 |
| ステータス | 承認済み（有効・制限付き） |
| 関連ADR | [ADR-002](002-evaluation-strategy.md), [ADR-006](006-tco-trampoline.md), [ADR-008](008-dynamic-wind-limitations.md) |
| 関連PR | [PR #9](https://github.com/suchi/sucheme-go/pull/9) |

## コンテキスト

R7RS は `call-with-current-continuation`（`call/cc`）で取得した継続を任意の回数呼び出せることを要求する（re-entrant continuation / full continuation）。しかし、Goのランタイムはスタックフレームのコピーや復元をネイティブにサポートしていない。

### 検討した選択肢

1. **Escape continuation のみ**: `panic`/`recover` を使い、現在の計算を「脱出」する一方向の継続のみサポート
2. **CPS変換**: 評価器全体をCPS（継続渡しスタイル）に変換し、継続をGoのクロージャとして表現
3. **スタックコピー方式**: Goのgoroutineスタックをコピー・復元する（unsafe操作が必要）
4. **Setjmp/longjmp相当**: cgoを使用してC言語のsetjmp/longjmpを呼び出す

## 決定

**Escape continuation のみをサポートする。** `panic`/`recover` を使った脱出機構で実装する。

## 理由

1. **実装の現実性**: CPS変換は評価器の全面的な書き換えが必要で、現段階ではコスト対効果が見合わない
2. **多くのユースケースをカバー**: 非ローカル脱出（`guard`の例外処理、ループ脱出、コルーチンのキャンセル等）はescape continuationで十分
3. **安全性**: `unsafe` パッケージやcgoへの依存を避けられる
4. **WASMとの互換性**: WASMビルドでも問題なく動作する
5. **段階的な拡張パス**: 将来的にCPS変換への移行は可能（[ロードマップ「3. Full Continuations」](../enhancement-roadmap.md)）

## 実装（擬似コード）

実際の実装は `pkg/evaluator/continuation.go` を参照。以下は概要を示す擬似コード:

```go
type ContinuationEscape struct {
    Value *types.Value
    Tag   int64  // 一意なタグで正しいcall/ccにマッチ
}

// call/cc 内部:
// 1. ユニークなタグを生成
// 2. 継続オブジェクト（呼び出されるとpanic）を作成
// 3. defer + recover でタグが一致するpanicをキャッチ
// 4. 一度使用されたらaliveフラグをfalseに設定
```

## 制約と影響

- **再入不可**: 一度脱出した継続は再度呼び出せない（`alive`フラグで検出しエラーを返す）
- **`dynamic-wind` への影響**: ADR-008 の制限に直結。脱出時の `after` サンクは実行されるが、再入時の `before` サンクは実行されない
- **R7RS準拠表の記載**: `call/cc` は「部分対応」として記録
- **影響を受ける機能**: コルーチン、非決定性計算（`amb`）、マルチショット継続は実装不可

## 将来の変更

Full continuation を実装する場合は、ADR-002（評価戦略）の再検討が必要。CPS変換ベースの評価器への移行が最も現実的なパス。この場合、本ADRは **「ADR-007-v2」として更新** するか、新しいADRで本ADRを **「無効化（Superseded by ADR-XXX）」** する。
