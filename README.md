# DuckDB Go Example

このプロジェクトは、Go言語でDuckDBを使用する基本的なサンプルコードです。

## 必要条件

- Go 1.21以上
- DuckDB
- Vector Similarity Search Extension
- OpenAI APIキー

## セットアップ

1. 依存関係のインストール:
```bash
go mod download
```

2. 環境変数の設定:
`.env`ファイルを作成し、以下のように設定します：
```bash
OPENAI_API_KEY=your-api-key-here
```

3. サンプルドキュメントの準備:
`data`ディレクトリに検索対象のmarkdownファイルを配置します。

## 使用方法

### データベースの作成

データベースとサンプルデータを作成します：

```bash
go run create_db.go
```

### テキスト検索

テキストを指定して類似度検索を実行します：

```bash
go run search.go -text "検索したいテキスト"
```

## 機能

- DuckDBデータベースの作成
- Vector Similarity Search Extensionの使用
- OpenAI Embedding APIを使用したテキストのベクトル化
- コサイン類似度を使用した類似度検索
- Markdownファイルからのテキスト抽出

## スクリプトの説明

### create_db.go

このスクリプトでは以下の操作を行います：

1. `.env`ファイルから環境変数を読み込み
2. DuckDBデータベース（vectors.db）の作成
3. Vector Similarity Search Extensionのロード
4. `data`ディレクトリからmarkdownファイルを読み込み
5. OpenAI Embedding APIを使用してテキストをベクトル化
6. テキストとベクトルデータの格納

### search.go

このスクリプトでは以下の操作を行います：

1. `.env`ファイルから環境変数を読み込み
2. コマンドライン引数から検索テキストを取得
3. OpenAI Embedding APIを使用して検索テキストをベクトル化
4. DuckDBデータベースへの接続
5. Vector Similarity Search Extensionのロード
6. 類似度検索の実行と結果の表示

## ベクトル検索について

このサンプルでは、OpenAIのEmbedding APIを使用してテキストを1536次元のベクトルに変換します：

- ベクトルは `FLOAT[1536]` 型で格納されます
- コサイン類似度（`<=>` 演算子）を使用して類似度を計算します
- 類似度は0から1の間の値で、1に近いほど類似度が高いことを示します 