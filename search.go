package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

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

func main() {
	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// コマンドライン引数の解析
	searchText := flag.String("text", "", "検索するテキスト")
	flag.Parse()

	if *searchText == "" {
		log.Fatal("検索テキストを指定してください。例: -text \"検索したいテキスト\"")
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

	// クエリテキストからベクトルを生成
	queryVec, err := getEmbedding(client, *searchText)
	if err != nil {
		log.Fatalf("クエリテキストのベクトル生成に失敗しました: %v", err)
	}

	// ベクトルを文字列に変換
	vecStr := "["
	for i, v := range queryVec {
		if i > 0 {
			vecStr += ", "
		}
		vecStr += fmt.Sprintf("%f", v)
	}
	vecStr += "]"

	// コサイン類似度を使用した類似度検索
	rows, err := db.Query(`
		SELECT id, text, 
		       1 - (embedding <=> ?) as similarity
		FROM documents
		ORDER BY similarity DESC
		LIMIT 3
	`, vecStr)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("検索テキスト: \"%s\"\n", *searchText)
	fmt.Println("類似度検索結果:")
	for rows.Next() {
		var id int
		var text string
		var similarity float64
		err := rows.Scan(&id, &text, &similarity)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, テキスト: %s, 類似度: %.4f\n", id, text, similarity)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}
