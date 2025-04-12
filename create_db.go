package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/sashabaranov/go-openai"
)

func getEmbedding(client *openai.Client, text string) ([]float32, error) {
	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2,
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

func readMarkdownFiles(dir string) ([]string, error) {
	var texts []string

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			texts = append(texts, string(content))
		}
	}

	return texts, nil
}

func main() {
	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// OpenAI APIキーの取得
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY環境変数が設定されていません")
	}

	// OpenAIクライアントの初期化
	client := openai.NewClient(apiKey)

	// DuckDBデータベースに接続
	db, err := sql.Open("duckdb", "document.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Vector Similarity Search Extensionをロード
	_, err = db.Exec("INSTALL 'vss';")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("LOAD 'vss';")
	if err != nil {
		log.Fatal(err)
	}

	// ベクトルテーブルの作成
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id INTEGER,
			text VARCHAR,
			embedding FLOAT[1536]
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// サンプルテキストとベクトルデータの生成と挿入
	stmt, err := db.Prepare("INSERT INTO documents (id, text, embedding) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// dataディレクトリからmarkdownファイルを読み込む
	texts, err := readMarkdownFiles("data")
	if err != nil {
		log.Fatal(err)
	}

	if len(texts) == 0 {
		log.Fatal("dataディレクトリにmarkdownファイルが見つかりません")
	}

	for i, text := range texts {
		// テキストを整形（改行をスペースに置換）
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.TrimSpace(text)

		// OpenAIのEmbedding APIを使用してベクトルを生成
		embedding, err := getEmbedding(client, text)
		if err != nil {
			log.Fatalf("テキストのベクトル生成に失敗しました: %v", err)
		}

		// ベクトルを文字列に変換
		vecStr := "["
		for j, v := range embedding {
			if j > 0 {
				vecStr += ", "
			}
			vecStr += fmt.Sprintf("%f", v)
		}
		vecStr += "]"

		// データベースに挿入
		_, err = stmt.Exec(i+1, text, vecStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("データベースの作成が完了しました。%d件のドキュメントを登録しました。\n", len(texts))
} 