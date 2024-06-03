package main

import (
    "database/sql"
    "log"
    "fmt"
    "time"
    "math/rand"
    "net/http"

    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func PopulateUsers() error {
    // Prepare the SQL statement
    stmt, err := db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    // Seed random number generator
    rand.Seed(time.Now().UnixNano())

    // Characters for generating random password
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    // Loop to insert 5000 users
    for i := 1; i <= 5000; i++ {
        // Generate username
        username := fmt.Sprintf("RemoteVoice%04d", i)

        // Generate random password
        password := ""
        for j := 0; j < 8; j++ {
            password += string(chars[rand.Intn(len(chars))])
        }

        // Execute the SQL statement to insert user
        _, err := stmt.Exec(username, password)
        if err != nil {
            return err
        }
    }

    return nil
}

func AddUser(username, password string) error {
    // Prepare the SQL statement
    stmt, err := db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    // Execute the SQL statement
    _, err = stmt.Exec(username, password)
    if err != nil {
        return err
    }

    return nil
}

func addUserHandler(c *gin.Context) {
    // Bind JSON payload to the struct
    var json struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    if err := c.ShouldBindJSON(&json); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Add the user to the database
    err := AddUser(json.Username, json.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "User added successfully"})
}

func main() {
    // Connect to the MySQL database
    var err error
    db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/test_logins")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Ping the database to ensure a connection
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    // // Call the PopulateUsers function to add 5000 users
    // error := PopulateUsers()
    // if error != nil {
    //     log.Fatal(error)
    // }

    // Create a Gin router
    r := gin.Default()

    // // Hardcoded credentials
    // const (
    //     username = "adi"
    //     password = "password123"
    // )

    // Get the current local time
    currentTime := time.Now()
    
    // Define a route for the ping endpoint
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
            "message2": currentTime,
        })
    })

    r.POST("/login", func(c *gin.Context) {
        var json struct {
            Username string `json:"username" binding:"required"`
            Password string `json:"password" binding:"required"`
        }

        // Bind JSON payload to the struct
        if err := c.ShouldBindJSON(&json); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        // Check the credentials
        var storedPassword string
        err := db.QueryRow("SELECT password FROM users WHERE username = ?", json.Username).Scan(&storedPassword)
        if err != nil {
            if err == sql.ErrNoRows {
                c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        // Compare the stored password with the provided password
        if json.Password == storedPassword {
            c.JSON(http.StatusOK, gin.H{"status": "login successful"})
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
        }
    })

    // Define a route to add a user
    r.POST("/adduser", addUserHandler)

    // Start the server on port 8080
    r.Run(":8080")
}