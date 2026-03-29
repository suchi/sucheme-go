# CI/CD 修正計画

| 項目 | 値 |
|------|-----|
| 作成日 | 2026-03-28 |
| 完了日 | 2026-03-28 |
| ステータス | 解決済み |
| 関連PR | #12 (R7RS library system, CI/CD fix) |

## 問題

全GitHub Actions CI/CDの `build-wasm` ジョブが失敗している。

## 原因分析

1. **Go バージョン**: `go-version: '1.24.7'` が GitHub Actions の `setup-go@v5` で利用不可能な可能性
   - ローカル環境では Go 1.24.7 がインストールされている
   - GitHub Actions の Go ツールキャッシュに 1.24.7 が含まれていない可能性
   - 解決策: `go-version: '1.24'` に変更（マイナーバージョン内の最新パッチを自動選択）

2. **WASM ビルド**: `GOOS=js GOARCH=wasm go build` がCIで失敗
   - ローカルでは成功するため、Go バージョン問題が根本原因の可能性

## 修正計画

### 1. Go バージョンの修正

全ワークフロー（ci.yml, pages.yml, release.yml）の `go-version` を `'1.24'` に変更。
`setup-go@v5` はマイナーバージョン指定で最新パッチを自動選択する。

### 2. go.mod の整合性

`go.mod` の `go 1.24` は Go ツールチェーンの要求を示す。
CI が `1.24.x` (x < 7) をインストールした場合でもビルドは成功するはず。

### 3. CLAUDE.md ルール追加

> CI が全てグリーン（全ジョブ成功）でなければマージ禁止

## 確認手順

1. feature/libraries ブランチで修正
2. プッシュしてCI実行
3. 全ジョブが成功することを確認
4. Copilot レビュー完了後にマージ
