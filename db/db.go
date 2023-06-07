package db

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
)

// Error types
var (
	ErrWrongPassword = fmt.Errorf("wrong password")
)

type Task struct {
	ID          int    `db:"id" json:"id"`
	Description string `db:"description" json:"description"`
	Status      bool   `db:"status" json:"status"`
}

type User struct {
	ID       string `db:"id"`
	Login    string `db:"login" json:"login"`
	Password string `db:"password" json:"password"`
}

type DB struct {
	conn *sqlx.DB
}

func (d *DB) Auth(user User) (User, error) {
	var hash string
	if err := d.conn.Get(&hash, `
		SELECT password 
		FROM users
		WHERE login=$1`, user.Login); err != nil {
		return User{}, fmt.Errorf("user auth: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return User{}, ErrWrongPassword
		}
	}

	return User{
		Login: user.Login,
	}, nil
}

func (d *DB) CreateUser(user User) (User, error) {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generating hash: %w", err)
	}
	var id string
	if err := d.conn.Get(&id, `INSERT INTO users(login, password) VALUES ($1, $2) RETURNING id`, user.Login, string(hashpass)); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	user.ID = id
	return user, nil
}

func NewDB() (*DB, error) {
	connUri := "postgres://gleb:@localhost:5432/todo"

	//connection pull initialised
	db, err := sqlx.Connect("pgx", connUri)
	if err != nil {
		return nil, fmt.Errorf("sqlx connect: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(3)

	return &DB{
		conn: db,
	}, nil
}

//func (d *DB) SelectTasks() ([]Task, error) {
//	var tasks []Task
//	if err := d.conn.Select(&tasks, `SELECT id, description, status FROM tasks`); err != nil {
//		return nil, fmt.Errorf("select tasks: %w", err)
//	}
//
//	return tasks, nil
//}

//func (d *DB) GetTaskById(id string) (Task, error) {
//	var t Task
//	if err := d.conn.Get(&t, `SELECT id, description, status FROM tasks WHERE id =$1`, id); err != nil {
//		return Task{}, fmt.Errorf("select task %s: %w", id, err)
//	}
//
//	return t, nil
//}
