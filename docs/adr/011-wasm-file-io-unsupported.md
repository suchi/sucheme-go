# ADR-011: WASM ビルドでファイルI/Oを非対応とする

| 項目 | 値 |
|------|-----|
| 決定日 | 2026-03-28 |
| ステータス | 承認済み（有効） |
| 関連ADR | [ADR-001](001-use-go-for-scheme-interpreter.md)（GoのWASMサポートを選定理由に含む） |
| 関連PR | [PR #8](https://github.com/suchi/sucheme-go/pull/8)（ポートI/O実装）, [PR #26](https://github.com/suchi/sucheme-go/pull/26)（include実装） |

## コンテキスト

Sucheme は `GOOS=js GOARCH=wasm` でWASMにコンパイルし、ブラウザ上で動作する。しかし、ブラウザ環境にはネイティブなファイルシステムが存在しない。R7RSは `open-input-file`, `open-output-file`, `include`, `include-ci` 等のファイルI/O機能を定義している。

### 検討した選択肢

1. **WASMではファイルI/Oを非対応にする**: ビルドタグで分離し、WASMビルドではエラーを返す
2. **仮想ファイルシステムを実装**: メモリ上にファイルシステムをエミュレート
3. **File System Access API を使用**: ブラウザの File System Access API 経由でローカルファイルにアクセス

## 決定

**選択肢1（WASMではファイルI/Oを非対応にする）を採用する。** Goのビルドタグ（`//go:build !js` / `//go:build js`）でCLIビルドとWASMビルドを分離する。

## 理由

1. **シンプルさ**: ビルドタグによる分離は既存のGoの仕組みに乗るだけで実装コストがほぼゼロ
2. **セキュリティ**: ブラウザ上で任意のファイルへのアクセスを許可するのはセキュリティリスク
3. **仮想ファイルシステムのコスト**: 教育用途の処理系としては過剰な複雑さ
4. **WASM利用の主目的**: Web REPLでの対話的な使用が中心であり、ファイルI/Oの需要は低い

## 実装

```go
// pkg/evaluator/include.go (CLIビルド)
//go:build !js
// → os.ReadFile() でファイルを読み込み

// pkg/evaluator/include_wasm.go (WASMビルド)
//go:build js
// → "include: not supported in WASM builds" エラーを返す
```

同様のパターンは `open-input-file` / `open-output-file` にも適用。

## 影響

- R7RS準拠表の `include`, `include-ci` に「WASM非対応」の備考を付記
- `open-input-file`, `open-output-file` はR7RS準拠表で「未対応」（CLIでも未実装、[ロードマップ E-05](../enhancement-roadmap.md) で対応予定）
- Web UIのREPLではファイル関連機能が使えないことをユーザーに通知（`:help` コマンド等で案内）

## 将来の変更

CLIビルドでの `open-input-file` / `open-output-file` 実装（ロードマップ E-05）は本ADRの範囲外。WASMビルドでの仮想ファイルシステム導入を検討する場合は、新しいADRで判断を記録し、本ADRのステータスを「**一部無効化（Partially superseded by ADR-XXX）**」に変更する。
