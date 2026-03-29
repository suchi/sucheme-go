# Sucheme フィードバック対応プラン & SOW

## 作成日: 2026-03-29
## ステータス: 実行中

---

## SOW (Statement of Work)

### プロジェクト概要
ユーザフィードバックに基づく Sucheme の品質改善・機能追加プロジェクト。
CLI/Web REPL のキーバインド対応、サンプル品質向上、マルチプラットフォーム対応、
R7RS 準拠強化、テスト拡充、ダッシュボード改善、将来ロードマップ策定を含む。

### スコープ

| # | カテゴリ | フィードバック内容 | 対応方針 |
|---|---------|-------------------|---------|
| FB-1 | REPL キーバインド | CLI/Web で Emacs/vi キーバインドを使いたい | CLI: readline ライブラリ導入、Web: JS で実装 |
| FB-2 | サンプル品質 | パディング（スタブ）サンプルが30個あり不誠実 | 全スタブを削除し、本物のアルゴリズム/テクニックに置換 |
| FB-3 | マルチプラットフォーム | Windows でビルドコマンドが間違っている | 全ドキュメントの Windows 対応確認・修正 |
| FB-4 | テスト品質 | ユニット/機能/統合/E2E テストが十分か | テストカバレッジ拡充、境界/特殊/エラー/複雑ケース追加 |
| FB-5 | Web UI | もう少しユーザフレンドリーな機能を | テーマ切替、履歴永続化、ヘルプ、シンタックスハイライト等 |
| FB-6 | ダッシュボード | カバレッジ統計をもっとわかりやすく | プログレスバー、円グラフ、R7RS 進捗率の可視化 |
| FB-7 | R7RS 準拠 | 完全準拠するための作業 | 未対応10項目・部分対応10項目の対応 |
| FB-8 | エンハンスロードマップ | 今後追加すべき機能を20件以上検討 | ロードマップドキュメント作成 |

### 成果物
- PR としてマージされたコード変更
- 更新されたドキュメント
- 将来のエンハンスロードマップ (`docs/enhancement-roadmap.md`)

### 品質基準
- 全 PR は GitHub Copilot レビューで指摘ゼロまで修正
- CI/CD 全ジョブがグリーン
- `go test ./...` 全パス
- Conventional Commits に準拠

---

## 実行プラン

### フェーズ構成

作業はファイルパスの衝突を避けるため、フェーズごとに並行実行可能な単位に分割する。

### Round 1: 独立した5タスクを並行実行

ファイルパスが完全に分離されているため同時実行可能。

#### Task 1-A: サンプル品質改善
- **ブランチ**: `fix/examples-quality`
- **対象ファイル**: `examples/` のみ
- **内容**:
  - 30個のパディングファイル（math/51-60, practical/71-80, puzzles/91-100）を削除
  - 本物の教育的価値のあるSchemeプログラムに置換
  - math: 行列演算、組合せ数、カタラン数、オイラーφ関数、素因数分解等
  - practical: RPN電卓、状態機械、ストリーム、メモ化、テストフレームワーク等
  - puzzles: N-Queens、ハノイの塔、迷路探索、ナップサック問題、順列生成等
- **見積もり**: 中規模

#### Task 1-B: マルチプラットフォームドキュメント修正
- **ブランチ**: `fix/cross-platform-docs`
- **対象ファイル**: `README.md`, `docs/onboarding.md`, `docs/tutorial.md`, `CLAUDE.md`
- **内容**:
  - CLI ビルドコマンドの Windows 対応（`.exe` 拡張子）
  - WASM ビルドの Windows 注意事項（Git Bash/WSL 必要）
  - パス記法・シェルコマンドの注意事項
  - Go のクロスプラットフォームビルド方法の説明
- **見積もり**: 小規模

#### Task 1-C: CLI readline/キーバインド対応
- **ブランチ**: `feat/cli-readline`
- **対象ファイル**: `cmd/sucheme/main.go`, `go.mod`, `go.sum`
- **内容**:
  - `github.com/chzyer/readline` ライブラリ導入
  - Emacs キーバインド（デフォルト）: Ctrl+A/E, Ctrl+K, Ctrl+P/N, Ctrl+R
  - vi キーバインド: `--vi` フラグまたは `SUCHEME_KEYMAP=vi` 環境変数
  - コマンド履歴の永続化（`~/.sucheme_history`）
  - 既存の括弧バランスチェック機能との統合
- **見積もり**: 中規模

#### Task 1-D: Web UI 機能強化
- **ブランチ**: `feat/web-ui-enhancements`
- **対象ファイル**: `web-ui/index.html`, `web-ui/app.js`, `web-ui/style.css`
- **内容**:
  - Emacs/vi キーバインドモード切替
  - ダーク/ライトテーマ切替（localStorage 保持）
  - コマンド履歴の localStorage 永続化
  - Clear Output / Copy ボタン
  - `:help`, `:examples` コマンド
  - フォントサイズ調整（Ctrl+Plus/Minus）
  - 軽量シンタックスハイライト
- **見積もり**: 大規模

#### Task 1-E: ダッシュボード改善
- **ブランチ**: `feat/dashboard-v2`
- **対象ファイル**: `web-ui/dashboard.html`, `.github/workflows/ci.yml`（最小限）
- **内容**:
  - パッケージ別カバレッジのプログレスバー表示
  - 全体カバレッジの CSS 円グラフ
  - R7RS 対応状況のビジュアル表示
  - テスト結果のパッケージ別テーブル
  - グリッドレイアウトでカード型UI
  - ページ内ナビゲーション
- **見積もり**: 中規模

### Round 2: R7RS 準拠強化（Round 1 完了後）

#### Task 2-A: R7RS 未対応・部分対応機能の実装
- **ブランチ**: `feat/r7rs-compliance`
- **対象ファイル**: `pkg/evaluator/`, `pkg/builtin/`, `docs/r7rs-compliance.md`
- **内容**:

  **未対応（10項目）→ 実装対象:**
  | 機能 | 難易度 | 優先度 |
  |------|--------|--------|
  | `define-record-type` | 高 | 高 |
  | `cond-expand` | 中 | 高 |
  | `define-values` | 低 | 中 |
  | `include` / `include-ci` | 中 | 中 |
  | `read` (S式パーサ入力) | 中 | 高 |
  | `string-set!` | 低 | 低 |
  | `open-input-file` / `open-output-file` | 中 | 中（CLIのみ） |
  | 有理数 | 高 | 低 |
  | 複素数 | 高 | 低 |

  **部分対応（10項目）→ 完全対応:**
  | 機能 | 修正内容 |
  |------|---------|
  | `case` | 高度なパターン対応 |
  | `call/cc` | escape continuation 以外は困難（設計上の制約） |
  | `dynamic-wind` | re-entrant 対応 |
  | `assoc` | 任意比較手続き引数 |
  | `member` | 任意比較手続き引数 |
  | `values` | 複数値対応 |
  | `read-char` 等 | `current-input-port` 実装 |
  | `write-string` | `current-output-port` 実装 |
  | `peek-char` | `current-input-port` 実装 |

- **見積もり**: 大規模（複数コミットに分割）

### Round 3: テスト品質強化（Round 2 完了後）

#### Task 3-A: テストカバレッジ拡充
- **ブランチ**: `test/comprehensive-coverage`
- **対象ファイル**: `pkg/*_test.go`
- **内容**:
  - 各パッケージの境界ケースを網羅
  - Round 2 で追加した R7RS 機能のテスト
  - エラーケース・型エラーのテスト
  - 統合テスト（複数機能の組合せ）
  - CLI の E2E テスト（`-e` フラグ等）
- **見積もり**: 大規模

### Round 4: エンハンスロードマップ（全作業完了後）

#### Task 4-A: 今後のエンハンス候補 20 件以上のロードマップ作成
- **ブランチ**: `docs/enhancement-roadmap`
- **対象ファイル**: `docs/enhancement-roadmap.md`
- **内容**: 実装済み機能を踏まえた将来拡張の提案
- **見積もり**: 小規模

---

## リスクと対策

| リスク | 影響 | 対策 |
|--------|------|------|
| サブエージェントのフリーズ | 作業遅延 | 60秒ごとに出力監視、2回無変化でタスク引取り |
| Copilot レビューの遅延 | マージ遅延 | 最大3回再リクエスト、各60秒待機 |
| CI 失敗 | マージブロック | 最優先で修正、他タスクより先に対応 |
| go.mod の衝突 | マージ競合 | readline 追加は独立ブランチで、マージ順を制御 |
| API 529 Overloaded | エージェントクラッシュ | リトライ + タスク引取り |

---

## 進捗管理

| Round | タスク | ブランチ | PR | 状態 |
|-------|--------|---------|-----|------|
| 1-A | サンプル品質 | fix/examples-quality | #21 | 完了 |
| 1-B | ドキュメント修正 | fix/cross-platform-docs | #18 | 完了 |
| 1-C | CLI readline | feat/cli-readline | #19 | 完了 |
| 1-D | Web UI 強化 | feat/web-ui-enhancements | #20 | 完了 |
| 1-E | ダッシュボード | feat/dashboard-v2 | #22 | 完了 |
| 2-A | R7RS 準拠 | feat/r7rs-compliance | #26 | 完了 |
| 3-A | テスト強化 | test/comprehensive-coverage | - | 待機中 |
| 4-A | ロードマップ | docs/enhancement-roadmap | - | 待機中 |
