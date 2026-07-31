package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "os"
    "log"


    _ "github.com/lib/pq"
)

type Books struct{
    db *sql.DB
}

func NewServer() *Books {
    // Сначала пробуем взять строку из окружения
    connStr := os.Getenv("DATABASE_URL")
    // Если её нет — используем локальную для разработки
    if connStr == "" {
        connStr = "user=postgres password=36863686 dbname=work_db sslmode=disable"
    }
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    return &Books{db: db}
}


func (b *Books) saveBooks(title string, author string, year int64, rating float64) error{
    _, err := b.db.Exec("INSERT INTO books_list1 (title, author, year, rating) VALUES ($1,$2,$3,$4)", title, author, year, rating,)
    return err
}

func (b *Books) add_book(w http.ResponseWriter, r *http.Request){
    value := r.URL.Query()

    title := value.Get("value")
    author := value.Get("author")
    yearstr:= value.Get("year")
    ratingstr := value.Get("rating")

    year, err := strconv.ParseInt(yearstr, 10, 64)
    if err!=nil{
        fmt.Fprintf(w, "year not int")
        return
    }

    rating, err := strconv.ParseFloat(ratingstr, 64)
    if err!= nil{
        fmt.Fprintf(w, "rating not float64")
    }

    err = b.saveBooks(title, author, year, rating)
    if err!=nil{
        fmt.Fprintf(w, "error on save: %v", err)
    }
    fmt.Fprintf(w, "success")
}

func (b *Books) get_books(w http.ResponseWriter,r *http.Request){
    rows, err:= b.db.Query(`
        SELECT id, title, author, year, rating, created_at FROM books_list1;
    `)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "books list: \n")
    for rows.Next(){
        var id int
        var title string
        var author string
        var year int
        var rating float64
        var created_at time.Time

        err := rows.Scan(&id, &title, &author, &year,&rating, &created_at)
        if err!=nil{
            continue
        }

        fmt.Fprintf(w, "id: %d, title: %v, author: %v, year: %d, rating: %.2f,  created_at: %v\n", id, title, author, year,rating, created_at.Format("2006-01-02 15:04:05"))
    }
}

func (b *Books) put_book(w http.ResponseWriter,r *http.Request){
    idstr := r.URL.Query().Get("id")
    ratingstr := r.URL.Query().Get("rating")

    rating, err:= strconv.ParseFloat(ratingstr, 64)
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    _, err= b.db.Query(`
        UPDATE books_list1 SET rating=$1 WHERE id = $2;
    `, rating, id)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "success, book id: %d", id)

}



func (b *Books) get_book(w http.ResponseWriter,r *http.Request){
    idstr := r.URL.Query().Get("id")
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    rows, err:= b.db.Query(`
        SELECT id, title, author, year, rating, created_at FROM books_list1 WHERE id =$1;
    `, id)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "book list: %d\n", id)
    for rows.Next(){
        var id int
        var title string
        var author string
        var year int
        var rating float64
        var created_at time.Time



        err := rows.Scan(&id, &title, &author, &year,&rating, &created_at)
        if err!=nil{
            continue
        }

        fmt.Fprintf(w, "id: %d, title: %v, author: %v, year: %d, rating: %.2f,  created_at: %v\n", id, title, author, year,rating, created_at.Format("2006-01-02 15:04:05"))
    }
}

func (b *Books) get_books_stats(w http.ResponseWriter,r *http.Request){

    rows, err:= b.db.Query(`
        SELECT * FROM books_list1
        ORDER BY rating DESC
        LIMIT 5;
    `)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "top list\n")
    for rows.Next(){
        var id int
        var title string
        var author string
        var year int
        var rating float64
        var created_at time.Time



        err := rows.Scan(&id, &title, &author, &year,&rating, &created_at)
        if err!=nil{
            continue
        }

        fmt.Fprintf(w, "id: %d, title: %v, author: %v, year: %d, rating: %.2f,  created_at: %v\n", id, title, author, year,rating, created_at.Format("2006-01-02 15:04:05"))
    }
}

func (b *Books) del_book(w http.ResponseWriter,r *http.Request){
    idstr := r.URL.Query().Get("id")
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    _, err = b.db.Query(`
        DELETE FROM books_list1 WHERE id =$1;
    `, id)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "deleted book list: %d\n", id)

}

func (b *Books) getAuthors(w http.ResponseWriter, r *http.Request) {
    rows, err := b.db.Query(`
        SELECT DISTINCT author
        FROM books_list1
        ORDER BY author
    `)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var authors []string
    for rows.Next() {
        var author string
        if err := rows.Scan(&author); err != nil {
            continue
        }
        authors = append(authors, author)
    }

    // Формируем ответ
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(authors)
}


func main(){
    // Проверка переменной окружения
    dbURL := os.Getenv("DATABASE_URL")
    fmt.Println("DATABASE_URL =", dbURL)
    if dbURL == "" {
        log.Fatal("DATABASE_URL не найдена")
    }

    server := NewServer()

    createTableSQL := `
    CREATE TABLE books_list1 (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    year INT,
    rating DECIMAL(3,2) CHECK (rating >= 0 AND rating <= 5),
    created_at TIMESTAMP DEFAULT NOW()
    );`

    _, err := server.db.Exec(createTableSQL)
    if err != nil {
        log.Fatal("Ошибка создания таблицы:", err)
    }
    fmt.Println("Таблица tasks_list проверена/создана")


    

    http.HandleFunc("/top_rate", server.get_books_stats)
    http.HandleFunc("/authors", server.getAuthors)
    http.HandleFunc("/get_book", server.get_book)
    http.HandleFunc("/del_book", server.del_book)
    http.HandleFunc("/put_book", server.put_book)
    http.HandleFunc("/get_books", server.get_books)
    http.HandleFunc("/new_book", server.add_book)
    fmt.Println("Сервер запущен на http://localhost:8080")
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Println("Сервер запущен на порту", port)
    http.ListenAndServe(":"+port, nil)

}
