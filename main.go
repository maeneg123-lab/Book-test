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
    "context"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"


    _ "github.com/lib/pq"
)

var jwtSecret = []byte("твой_секретный_ключ_для_jwt")

func generateToken(userID int) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })
    return token.SignedString(jwtSecret)
}

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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
            return
        }

        // Убираем "Bearer " из строки
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("неверный метод подписи")
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Неверный токен", http.StatusUnauthorized)
            return
        }

        // Извлекаем user_id из токена и добавляем в контекст запроса
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            userID := int(claims["user_id"].(float64))
            r = r.WithContext(context.WithValue(r.Context(), "user_id", userID))
        }

        next(w, r)
    }
}

func (b *Books) saveBooks(title string, author string, year int, rating float64, userID int) error {
    _, err := n.db.Exec(
        "INSERT INTO books_list1 (title, author, year, rating, user_id) VALUES ($1, $2, $3, $4, $5)",
        title,author, year, rating, userID,
    )
    return err
}

func (b *Books) add_book(w http.ResponseWriter, r *http.Request){
    
    value := r.URL.Query()

    title := value.Get("title")
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
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    rows, err:= b.db.Query(`
        SELECT id, title, author, year, rating, created_at FROM books_list1 WHERE user_id=$1;
    `, userID)
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
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    idstr := r.URL.Query().Get("id")
    ratingstr := r.URL.Query().Get("rating")

    rating, err:= strconv.ParseFloat(ratingstr, 64)
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    _, err= b.db.Query(`
        UPDATE books_list1 SET rating=$1 WHERE id = $2 AND user_id=$3;
    `, rating, id, userID)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "success, book id: %d", id)

}



func (b *Books) get_book(w http.ResponseWriter,r *http.Request){
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    idstr := r.URL.Query().Get("id")
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    rows, err:= b.db.Query(`
        SELECT id, title, author, year, rating, created_at FROM books_list1 WHERE id =$1 AND user_id =$2;
    `, id, userID)
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
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    rows, err:= b.db.Query(`
        SELECT * FROM books_list1 WHERE user_id =$1
        ORDER BY rating DESC
        LIMIT 5;
    `, userID)
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
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    idstr := r.URL.Query().Get("id")
    id,err:=strconv.ParseInt(idstr,10,64)
    if err!=nil{
        fmt.Fprintf(w, "error, id not int")
        return
    }

    _, err = b.db.Query(`
        DELETE FROM books_list1 WHERE id =$1 AND user_id=$2;
    `, id, userID)
    if err!=nil{
        fmt.Fprintf(w, "error: %v", err)
    }
    fmt.Fprintf(w, "deleted book list: %d\n", id)

}

func (b *Books) getAuthors(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        http.Error(w, "User not authorized", http.StatusUnauthorized)
        return
    }
    rows, err := b.db.Query(`
        SELECT DISTINCT author
        FROM books_list1 WHERE user_id = $1
        ORDER BY author
    `, userID)
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

func (n *Notes) register(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    password := r.URL.Query().Get("password")

    if username == "" || password == "" {
        http.Error(w, "Username and password are required", http.StatusBadRequest)
        return
    }

    // Хешируем пароль
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    // Сохраняем в БД
    _, err = n.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, hashedPassword)
    if err != nil {
        http.Error(w, "Username already exists", http.StatusConflict)
        return
    }

    fmt.Fprintf(w, "User registered successfully")
}

func (n *Notes) login(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    password := r.URL.Query().Get("password")

    if username == "" || password == "" {
        http.Error(w, "Username and password are required", http.StatusBadRequest)
        return
    }

    // Ищем пользователя в БД
    var userID int
    var hashedPassword string
    err := n.db.QueryRow("SELECT id, password FROM users WHERE username=$1", username).Scan(&userID, &hashedPassword)
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Проверяем пароль
    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Генерируем JWT-токен
    token, err := generateToken(userID)
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    // Отправляем токен в ответе
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": token})
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
    CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() 
    );`

    _, err := server.db.Exec(createTableSQL)
    if err != nil {
        log.Fatal("Ошибка создания таблицы:", err)
    }
    fmt.Println("Таблица tasks_list проверена/создана")

    tableUserCreate:=`
    CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );`

    _, err:= server.db.Exec(tableUserCreate) 
    if err!= nil{
        log.Fatal("ошибка создания таблицы user:", err) 
    }
    fmt.Println("таблица user  создана") 
    updateBook := `
    ALTER TABLE book_list1 ADD COLUMN user_id INT REFERENCES users(id);`

    _, err := server.db.Exec(updateBook) 
    if err!=nil{
        log.Fatal("ошибка обновления таблицы: ", err) 
    }
    fmt.Println("таблица обновлена") 

  
    

    http.HandleFunc("/top_rate", authMiddleware(server.get_books_stats)) 
    http.HandleFunc("/register", server.register)
    http.HandleFunc("/login", server.login)
    http.HandleFunc("/authors", authMiddleware(server.getAuthors)) 
    http.HandleFunc("/get_book", authMiddleware(server.get_book)) 
    http.HandleFunc("/del_book", authMiddleware(server.del_book)) 
    http.HandleFunc("/put_book", authMiddleware(server.put_book)) 
    http.HandleFunc("/get_books", authMiddleware(server.get_books)) 
    http.HandleFunc("/new_book", authMiddleware(server.add_book)) 
    fmt.Println("Сервер запущен на http://localhost:8080")
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Println("Сервер запущен на порту", port)
    http.ListenAndServe(":"+port, nil)

}
