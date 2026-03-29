# Sucheme 開発履歴

プロジェクトの開発過程を時系列で記録する。将来の開発者がどの順序で何が実装され、どのような判断がなされたかを追跡できるようにする。

各ドキュメントの詳細は個別ファイルを参照。PR番号のリンクからGitHub上でコード差分とレビューコメントを確認できる。

[PRs一覧]: https://github.com/suchi/sucheme-go/pulls?q=is%3Apr+is%3Amerged

---

## 2026-03-27: プロジェクト開始

- **初期スキャフォールド** — SOW（作業範囲記述書）、ADR（アーキテクチャ決定記録）、Go module を作成
  - [ADR-001](adr/001-use-go-for-scheme-interpreter.md): Go言語でSchemeインタプリタを実装する判断
  - [ADR-002](adr/002-evaluation-strategy.md): 評価戦略（ASTインタプリタ方式）
  - [ADR-003](adr/003-memory-management.md): メモリ管理（GoのGCに委譲）
- **Phase 1: 基盤実装** — Types/Lexer/Parser/CI/CD/Web UI（コミット `2c63269`）

## 2026-03-28: コア機能の実装（Phase 1〜4）

- [**PR #2**](https://github.com/suchi/sucheme-go/pull/2): 評価器、環境、組み込み手続きの実装
  - この時点で `docs/plans/architecture-plan.md` を作成（パッケージ構成、型システム、評価フローの設計）
  - [ADR-006](adr/006-tco-trampoline.md): TCOにトランポリン方式を採用
- [**PR #4**](https://github.com/suchi/sucheme-go/pull/4): R7RS `syntax-rules` マクロシステム
- [**PR #5**](https://github.com/suchi/sucheme-go/pull/5): 例外処理（`guard`/`raise`）、`case-lambda`、`map`/`filter`/`for-each`
- [**PR #6**](https://github.com/suchi/sucheme-go/pull/6), [**#7**](https://github.com/suchi/sucheme-go/pull/7): ドキュメント整備（CI/CD、組み込み関数一覧、R7RS対応表、オンボーディング、サンプル104件）
  - PR #6 は一度 revert されて #7 で再投入
- [**PR #8**](https://github.com/suchi/sucheme-go/pull/8): `parameterize`、ポートI/O、README更新（Phase 4a）
  - この時点で `docs/plans/phase4-plan.md` を作成（Phase 4 の依存関係と分割戦略）
- [**PR #9**](https://github.com/suchi/sucheme-go/pull/9): `call/cc`（escape continuation）、`dynamic-wind`、`call-with-values`（Phase 4b）
  - [ADR-005](adr/005-review-rule-enforcement.md): レビュールール強制の方針を策定
  - [ADR-007](adr/007-escape-continuation-only.md): `call/cc` を escape continuation のみで実装する判断
  - [ADR-008](adr/008-dynamic-wind-limitations.md): `dynamic-wind` の re-entrant 未対応の判断
- [**PR #10**](https://github.com/suchi/sucheme-go/pull/10): PR #7 レビュー対応 — 50件以上のエッジケーステスト追加
- [**PR #12**](https://github.com/suchi/sucheme-go/pull/12): R7RS ライブラリシステム（`define-library`/`import`）、CI/CD修正（Phase 4c）
  - この時点で `docs/plans/ci-fix-plan.md` を作成（`build-wasm` ジョブ失敗の原因分析と修正方針）
  - 問題: GitHub Actions の Go バージョン指定が厳密すぎた → `'1.24'` に変更

## 2026-03-29 午前: ドキュメント・ダッシュボード整備

- [**PR #13**](https://github.com/suchi/sucheme-go/pull/13): オンボーディングガイド書き直し、サブエージェントPRルール追加
- [**PR #14**](https://github.com/suchi/sucheme-go/pull/14): テスト/カバレッジダッシュボード + GitHub Pages デプロイ
- [**PR #15**](https://github.com/suchi/sucheme-go/pull/15): ダッシュボードレビュー対応（パッケージカバレッジ、ビューポート、クリーンアップ）
- [**PR #16**](https://github.com/suchi/sucheme-go/pull/16): 包括的な Scheme チュートリアル追加
- [**PR #17**](https://github.com/suchi/sucheme-go/pull/17): R7RS対応表更新（`define-library`/`import`対応）

## 2026-03-29 昼: フィードバック対応 Round 1（5タスク並行）

ユーザフィードバックに基づく品質改善。`docs/plans/feedback-plan-2026-03-29.md` に計画を策定し、サブエージェント5並列で実行。

- [**PR #18**](https://github.com/suchi/sucheme-go/pull/18): マルチプラットフォームドキュメント修正（Windows対応ビルドコマンド）
- [**PR #19**](https://github.com/suchi/sucheme-go/pull/19): CLI readline/キーバインド対応（Emacs/vi、履歴永続化）
- [**PR #20**](https://github.com/suchi/sucheme-go/pull/20): Web UI機能強化（テーマ切替、履歴永続化、シンタックスハイライト）
- [**PR #21**](https://github.com/suchi/sucheme-go/pull/21): サンプル品質改善（30個のスタブを本物のアルゴリズムに置換）
- [**PR #22**](https://github.com/suchi/sucheme-go/pull/22): ダッシュボードリデザイン（カバレッジ可視化、R7RS進捗バー）

## 2026-03-29 夕方: Round 2 — R7RS準拠強化

- [**PR #26**](https://github.com/suchi/sucheme-go/pull/26): R7RS特殊形式の大規模実装
  - 新規実装: `define-record-type`, `cond-expand`, `define-values`, `include`/`include-ci`, `case`の`=>`構文
  - 組み込み関数改善: `assoc`/`member`のオプショナル比較手続き、`read`、`string-set!`、ポート操作
  - R7RS対応率: 約80% → 90%超
  - [ADR-009](adr/009-include-ci-whole-source-lowercase.md): `include-ci` をソース全体小文字化で仮実装する判断
  - [ADR-009](adr/009-include-ci-whole-source-lowercase.md) 補足: PR #27 で「部分対応」に格下げ（レビュー指摘により非準拠が明確化）

## 2026-03-29 夜: Round 2.5 — CI修正・ドキュメント同期 + Round 3・4

- [**PR #27**](https://github.com/suchi/sucheme-go/pull/27): Dependabot cooldown構文修正、R7RS対応表・ダッシュボード数値同期
  - [ADR-010](adr/010-dependabot-cooldown-policy.md): Dependabot cooldown セキュリティポリシー策定
  - インシデント: Copilotレビューが `cooldown` を「無効なキー」と指摘 → 削除するミスが過去に発生。以後、削除禁止ルールをCLAUDE.mdに明記
- [**PR #28**](https://github.com/suchi/sucheme-go/pull/28): エンハンスロードマップ作成（30件の将来候補、6カテゴリ、Mermaid ganttタイムライン）
- [**PR #29**](https://github.com/suchi/sucheme-go/pull/29): テストカバレッジ大幅拡充（6パッケージ、2,618行追加、カバレッジ 68% → 80%）

---

## ドキュメントマップ

| ドキュメント | 作成日 | 種別 | 概要 |
|-------------|--------|------|------|
| [`docs/adr/001-*.md`](adr/001-use-go-for-scheme-interpreter.md) | 2026-03-27 | ADR | Go言語選定の判断 |
| [`docs/adr/002-*.md`](adr/002-evaluation-strategy.md) | 2026-03-27 | ADR | 評価戦略の判断 |
| [`docs/adr/003-*.md`](adr/003-memory-management.md) | 2026-03-27 | ADR | メモリ管理の判断 |
| [`docs/adr/005-*.md`](adr/005-review-rule-enforcement.md) | 2026-03-28 | ADR | レビュールール強制の判断 |
| [`docs/adr/006-*.md`](adr/006-tco-trampoline.md) | 2026-03-28 | ADR | TCOトランポリン方式の判断 |
| [`docs/adr/007-*.md`](adr/007-escape-continuation-only.md) | 2026-03-28 | ADR | escape continuation限定の判断 |
| [`docs/adr/008-*.md`](adr/008-dynamic-wind-limitations.md) | 2026-03-28 | ADR | dynamic-wind制限の判断 |
| [`docs/adr/009-*.md`](adr/009-include-ci-whole-source-lowercase.md) | 2026-03-29 | ADR | include-ci仮実装の判断 |
| [`docs/adr/010-*.md`](adr/010-dependabot-cooldown-policy.md) | 2026-03-29 | ADR | Dependabot cooldownポリシー |
| [`docs/adr/011-*.md`](adr/011-wasm-file-io-unsupported.md) | 2026-03-28 | ADR | WASMファイルI/O非対応の判断 |
| [`docs/plans/architecture-plan.md`](plans/architecture-plan.md) | 2026-03-28 | 設計 | パッケージ構成・型システム・評価フロー |
| [`docs/plans/phase4-plan.md`](plans/phase4-plan.md) | 2026-03-28 | 計画 | Phase 4（高度な機能）の実装戦略 |
| [`docs/plans/ci-fix-plan.md`](plans/ci-fix-plan.md) | 2026-03-28 | 計画 | CI/CD build-wasm失敗の修正 |
| [`docs/r7rs-compliance.md`](r7rs-compliance.md) | 2026-03-28 | 状態 | R7RS準拠状況（随時更新） |
| [`docs/onboarding.md`](onboarding.md) | 2026-03-28 | ガイド | 開発者向けオンボーディング |
| [`docs/tutorial.md`](tutorial.md) | 2026-03-29 | ガイド | Schemeチュートリアル |
| [`docs/plans/feedback-plan-2026-03-29.md`](plans/feedback-plan-2026-03-29.md) | 2026-03-29 | 計画 | フィードバック対応のSOW・実行プラン |
| [`docs/enhancement-roadmap.md`](enhancement-roadmap.md) | 2026-03-29 | 計画 | 将来のエンハンス候補30件 |
| [`docs/CHANGELOG.md`](CHANGELOG.md) | 2026-03-29 | 履歴 | **本ドキュメント** |
