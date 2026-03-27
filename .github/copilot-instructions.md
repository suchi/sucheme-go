# Copilot / Claude Code / Codex 共通インストラクション

## プロジェクト概要

**Sucheme** は Go で実装する R7RS 準拠の Scheme インタプリタです。
リポジトリ名の "sucheme" はオーナー suchi の名前に由来します（schemeではなくsucheme）。

## 言語規約

- 応答・説明・PR・ドキュメント (*.md): **日本語**
- コミットメッセージ: **英語**
- ソースコード内のログ・エラーメッセージ・UIラベル: **英語**
- ドキュメントには理解を補助するための Mermaid 図を付ける（ASCII図は使わない）

## 開発方針

- **テスト駆動開発 (TDD)**: テストを先に書いてから実装する
- **こまめなコミット**: 節目ごとにコミットしてブランチにプッシュ
- **サブエージェント活用**: 独立した作業は並行で進める
- **小さなPR**: タスクごとにスコープを絞る
- **CIグリーン必須**: コミット前にテスト実行

## コーディング規約

- Go の標準スタイルに従う (`gofmt`, `go vet`)
- 不要な機能・抽象化は追加しない
- エラーメッセージは英語で意味のある内容にする
- 秘密情報はハードコードしない

## アーキテクチャ

```
sucheme-go/
├── cmd/sucheme/     # CLI エントリポイント
├── pkg/
│   ├── lexer/       # トークナイザ
│   ├── parser/      # S式パーサ
│   ├── evaluator/   # 評価器
│   ├── types/       # Scheme値の型定義
│   ├── env/         # 環境（スコープ）
│   ├── builtin/     # 組み込み関数
│   └── macro/       # マクロ展開器
├── web-ui/          # Web UI (GitHub Pages)
├── examples/        # サンプル Scheme プログラム
├── docs/            # ドキュメント
│   ├── adr/         # Architecture Decision Records
│   ├── tutorial/    # チュートリアル
│   └── r7rs/        # R7RS 対応状況
└── benchmarks/      # ベンチマーク
```

## テスト

- `go test ./...` ですべてのテストを実行
- カバレッジ: `go test -coverprofile=coverage.out ./...`
- ベンチマーク: `go test -bench=. ./...`

## CI/CD

- GitHub Actions で Linux/macOS/Windows + WASM のビルド・テスト
- GitHub Pages で Web UI デプロイ
- Release 時にバイナリ配布
- PR 時にプレビューデプロイ

## パフォーマンス

- ベンチマークに竹内関数 (Takeuchi function) を含める
- CI で毎回パフォーマンス測定

## ライセンス

NOLICENSE（ライセンスなし）

## PR・レビュー・マージ

1. 各成果物はPRとして作成する
2. GitHub Copilotにレビューを依頼する
3. レビュー指摘はすべて解決する
4. CopilotレビューがNo New Commentになったら自動でmainにマージする

## 対応プラットフォーム

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
- WebAssembly (js/wasm)
