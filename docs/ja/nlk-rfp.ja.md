# RFP: nlk

> Generated: 2026-04-11
> Status: Implemented (v0.3.2)
>
> **Note:** This document records the initial planning discussion. The final
> implementation differs in several areas due to design decisions made during
> development. See "Implementation Notes" at the end for a summary of changes.

## 1. Problem Statement

nlink-jpの複数LLMアプリケーション（mail-analyzer, gem-cli, ai-ir2, news-collector等）が、LLM SDK周辺の共通機能（プロンプトインジェクションガード、LLM出力のJSON修復、出力検証、リトライ用バックオフ計算等）をそれぞれ独自に実装しており、類似コードが散在している。これらをGo/Python双方で利用可能な軽量共通ライブラリに集約する。LLM呼び出し自体の抽象化は行わず、呼び出しの「周辺」に特化したツールボックスとして提供する。フレームワークではない。

## 2. Functional Specification

### Commands / API Surface

CLIツールではなくGoライブラリ（`github.com/nlink-jp/nlk`）。以下の5パッケージを提供する。

| パッケージ | 機能 | 性質 |
|-----------|------|------|
| `guard` | ノンスタグXMLラッピングによるプロンプトインジェクション防御 | 純粋なデータ変換 |
| `prompt` | プロンプトテンプレート構築 | 純粋なデータ変換 |
| `jsonfix` | 不正JSON抽出・修復・パース | 純粋なデータ変換 |
| `validate` | バリデーション関数の実行フレームワーク（ルールはアプリが定義） | 純粋な検証 |
| `backoff` | ジッター付き指数バックオフ待ち時間計算 | 純粋な計算 |

全パッケージが純粋関数・ステートレス。各パッケージは独立して利用可能。

### Input / Output

各パッケージは関数呼び出しで利用する。stdin/stdout は関与しない。

```go
// guard: 非信頼データをラッピング
guarded := guard.Wrap(untrustedData)

// jsonfix: 不正JSONを修復
parsed, err := jsonfix.Parse(rawLLMOutput)

// backoff: 待ち時間計算
time.Sleep(backoff.Duration(attempt))

// validate: アプリ定義のルールで検証
err := validate.Run(result, myRules...)

// prompt: テンプレートからプロンプト構築
text := prompt.Build(template, vars)
```

### Configuration

設定ファイルや環境変数は不要。全てが関数引数で制御される。

### External Dependencies

なし。全パッケージがGo標準ライブラリのみで実装される。

## 3. Design Decisions

### 技術スタック

- **Go**: 初期実装言語。nlink-jpのLLMツールの多数がGoで実装されているため
- **Python版**: 後続で別リポジトリ（`nlk-py`）として開発予定

### 設計原則

- **ツールボックスであってフレームワークではない** — 呼び出しフローを規定しない。アプリが自由に組み合わせる
- **LLM API抽象化は行わない** — Vertex AIとOpenAI互換APIの差異を無理に統一しない。LLM呼び出し自体はアプリの責務
- **外部依存ゼロ** — サプライチェーン攻撃対策。Go標準ライブラリのみで実装
- **純粋関数・ステートレス** — 全パッケージが副作用を持たない

### 既存ツールとの関係

- `guard`: gem-cli, mail-analyzerのインジェクションガードから抽出
- `jsonfix`: json-filterのJSON抽出・修復ロジックを移植
- `backoff`: mail-analyzer, gem-cliのリトライロジックから抽出
- `validate`: mail-analyzerの出力検証パターンを汎用化
- `prompt`: 各ツールのプロンプト構築パターンを共通化

### LLMアプリケーションワークフロー上の位置づけ

```
  ユーザー入力 / 外部データ
        │
        ▼
  ┌─────────┐
  │  guard   │  前処理: 非信頼データをラッピング
  └────┬────┘
       ▼
  ┌─────────┐
  │  prompt  │  前処理: プロンプト組み立て
  └────┬────┘
       ▼
    LLM API      アプリ固有（Vertex AI / OpenAI互換 / ローカル）
       │         backoff.Duration() で待ち時間計算、ループはアプリ側
       ▼
  ┌─────────┐
  │ jsonfix  │  後処理: 生レスポンスからJSON抽出・修復
  └────┬────┘
       ▼
  ┌──────────┐
  │ validate  │  後処理: アプリ定義ルールで検証
  └────┬─────┘
       ▼
  アプリケーションロジック
```

### スコープ外

- LLM API呼び出しの抽象化
- パイプライン/ワークフローエンジン
- リトライループの制御（backoffは待ち時間計算のみ）
- Python版（別リポジトリとして後続）
- JSON Schemaバリデーション（validateはアプリ定義関数の実行フレームワーク）

## 4. Development Plan

### Phase 1: Core

- `guard` パッケージ実装 + テスト
- `jsonfix` パッケージ実装 + テスト（json-filterからのロジック移植）
- `backoff` パッケージ実装 + テスト
- 基盤ライブラリのため高テストカバレッジ必須

### Phase 2: Features

- `prompt` パッケージ実装 + テスト
- `validate` パッケージ実装 + テスト
- 既存ツール1つ（mail-analyzer等）をnlkに移行して実証

### Phase 3: Release

- ドキュメント（README.md, docs/ja/README.md, CHANGELOG.md）
- AGENTS.md
- Go module公開（タグ付け）

### レビュー単位

各Phaseを独立してレビュー可能。

## 5. Required API Scopes / Permissions

None。ライブラリ自体は外部APIを呼び出さない。

## 6. Series Placement

Series: **lib-series**（新設）
Reason: 全シリーズ横断で使われる共通基盤ライブラリ。特定シリーズに属するのは不適切。ライブラリ専用の新シリーズを設ける。

## 7. External Platform Constraints

None。純粋なGoライブラリであり、外部プラットフォームへの依存はない。

---

## Discussion Log

1. **課題認識**: 多数のLLMアプリが類似のSDKラッパーとインジェクションガードを個別実装している現状を改善したい
2. **スコープ決定**: LLM呼び出し自体の抽象化は行わない。Vertex AIとOpenAI互換APIの差異を無理に統一すると既存SDKの二番煎じになる。呼び出しの「周辺」に特化
3. **言語**: Go/Python両対応を目指すが、まずGoから着手
4. **retryの検討**: 当初retryモジュールを検討 → パイプライン化の提案 → フレームワーク化のリスクを認識 → 指数バックオフ計算のみを `backoff` として提供する方針に落ち着く
5. **外部依存ポリシー**: 全パッケージ外部依存ゼロを実現。サプライチェーン攻撃対策として重要。依存が必要になった場合は明示的に宣言する方針
6. **validateの設計**: JSON Schema準拠の汎用バリデーションではなく、アプリがバリデーション関数を渡す仕組みだけ提供（C案）。これにより外部依存ゼロを維持
7. **命名**: `nlk`（nlink-jpの略）を採用。名前空間衝突を防ぎつつ簡潔
8. **シリーズ**: 全シリーズ横断の基盤ライブラリのため、既存シリーズには属さず `lib-series` を新設

---

## Implementation Notes

以下は、RFP策定後の実装過程で変更された点のまとめ。
RFP本文は計画時点の議論記録としてそのまま残している。

### パッケージ構成の変更

| RFP計画 | 最終実装 | 理由 |
|---------|---------|------|
| `prompt` | **未実装** | 既存ツールの調査で、guard.Expand() + fmt.Sprintf で十分と判断。専用パッケージは過剰 |
| — | `strip` **追加** | LLM thinking/reasoningタグ除去の需要。jsonfix内部ではなく独立パッケージとして提供（パッケージ間依存ゼロの原則） |

最終パッケージ: `guard`, `jsonfix`, `strip`, `backoff`, `validate`

### API名の変更

| RFP記載 | 実装 |
|---------|------|
| `jsonfix.Parse()` | `jsonfix.Extract()` — 「抽出+修復」の意図を明確化 |
| `validate.Run(result, rules...)` | `validate.Run(rules...)` — resultは各ルールのクロージャがキャプチャ |

### jsonfix実装方針の変更

RFPでは「json-filterからのロジック移植」としていたが、Python json-repair（MIT, Stefano Baccianella）の修復ヒューリスティクスを参考に再帰下降パーサーとしてGoでゼロから実装。正規表現ベースではReDoSリスクや対応ケースの限界があったため。

### セキュリティ強化

- guard: ノンスサイズを4バイト（32ビット）→16バイト（128ビット）に引き上げ。ブルートフォース推測防止
- backoff: 負数のattemptを0にクランプ
- jsonfix: 巨大入力に対するメモリ使用量の注記をドキュメントに追加

### ワークフロー図（最終版）

```
  ユーザー入力 / 外部データ
        │
        ▼
  ┌─────────┐
  │  guard   │  前処理: 非信頼データをラッピング
  └────┬────┘
       ▼
    LLM API      アプリ固有（Vertex AI / OpenAI互換 / ローカル）
       │         backoff.Duration() で待ち時間計算、ループはアプリ側
       ▼
  ┌─────────┐
  │  strip   │  後処理: thinking/reasoningタグ除去（ローカルLLM向け）
  └────┬────┘
       ▼
  ┌─────────┐
  │ jsonfix  │  後処理: 生レスポンスからJSON抽出・修復
  └────┬────┘
       ▼
  ┌──────────┐
  │ validate  │  後処理: アプリ定義ルールで検証
  └────┬─────┘
       ▼
  アプリケーションロジック
```
