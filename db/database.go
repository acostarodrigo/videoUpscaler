package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/janction/videoUpscaler/videoUpscalerLogger"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Task represents a video upscaler task.
type Task struct {
	TaskId           string
	ThreadId         string
	WorkerSubscribed bool
}

// thread represents a video upscaler thread.
type Thread struct {
	ID                  string
	DownloadStarted     bool
	DownloadCompleted   bool
	WorkStarted         bool
	WorkCompleted       bool
	SolutionProposed    bool
	VerificationStarted bool
	SolutionRevealed    bool
	SubmitionStarted    bool
}

type Worker struct {
	Address    string
	Registered bool
}

type LogEntry struct {
	ThreadId  string
	Log       string
	Timestamp int64
	Severity  int64
}

type IPFS struct {
	Address string
	Added   bool
}

// DB encapsulates the database connection.
type DB struct {
	conn *sql.DB
}

type Database interface {
	UpdateTask(taskId, threadId string, completed bool) error
	UpdateThread(id string, downloadStarted, downloadCompleted, workStarted, workCompleted, solProposed, verificationStarted, solutionRevealed bool, submitionStarted bool) error
	AddLogEntry(threadId, log string, timestamp, severity int64) error
}

// Init initializes the SQLite database and creates the threads table.
func Init(databasePath string) (*DB, error) {
	// if the path doesn't exists, it might be that client wasn't yet initialized, so we don't create it
	_, err := os.Stat(databasePath)
	if errors.Is(err, fs.ErrNotExist) {
		return &DB{}, nil
	}

	db, err := sql.Open("sqlite3", filepath.Join(databasePath, "videoUpscaler.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	createTables := `
    CREATE TABLE IF NOT EXISTS tasks (
        taskId TEXT PRIMARY KEY,
        threadId TEXT,
		worker_subscribed BOOLEAN
	);
    CREATE TABLE IF NOT EXISTS threads (
        id TEXT PRIMARY KEY,
		download_started BOOLEAN,
		download_completed BOOLEAN,
		work_started BOOLEAN,
		work_completed BOOLEAN,
		solution_proposed BOOLEAN,
		solution_revealed BOOLEAN,
		verification_started BOOLEAN,
		submition_started BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS workers (
		address TEXT PRIMARY KEY,
		registered BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS logs (
		threadId TEXT,
		log TEXT,
		timestamp NUMBER,
		severity NUMBER
	);
	CREATE TABLE IF NOT EXISTS ipfs (
		address TEXT PRIMARY KEY,
		added BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS render_times (
		thread_id TEXT ,
		frame_number NUMBER,
		render_duration NUMBER
	);
    `

	if _, err := db.Exec(createTables); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DB{conn: db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Createthread inserts a new thread into the database.
func (db *DB) AddTask(taskId, threadId string) error {
	insertQuery := `INSERT INTO tasks (taskId, threadId, worker_subscribed) VALUES (?,?,false)`
	_, err := db.conn.Exec(insertQuery, taskId, threadId)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// Readthread retrieves a thread by ID.
func (db *DB) ReadTask(taskId, threadId string) (*Task, error) {
	query := `SELECT taskId, threadId, worker_subscribed  FROM tasks WHERE taskId = ? AND threadId = ? `
	row := db.conn.QueryRow(query, taskId, threadId)

	var task Task
	if err := row.Scan(&task.TaskId, &task.ThreadId, &task.WorkerSubscribed); err != nil {
		if err == sql.ErrNoRows {
			// thead doesn't exists, so we insert it
			db.AddTask(taskId, threadId)
			return &Task{TaskId: taskId, ThreadId: threadId, WorkerSubscribed: false}, nil
		}
		return nil, fmt.Errorf("failed to read thread: %w", err)
	}

	return &task, nil
}

// Updatethread updates a task's information.
func (db *DB) UpdateTask(taskId, threadId string, workerSubscribed bool) error {
	updateQuery := `UPDATE tasks SET worker_subscribed = ? WHERE taskId = ? AND threadId = ? `
	_, err := db.conn.Exec(updateQuery, workerSubscribed, taskId, threadId)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// Createthread inserts a new thread into the database.
func (db *DB) AddThread(id string) error {
	insertQuery := `INSERT INTO threads (id, download_started, download_completed, work_started, work_completed, solution_proposed,verification_started, solution_revealed, submition_started) VALUES (?, false, false, false, false, false, false, false, false)`
	_, err := db.conn.Exec(insertQuery, id)
	if err != nil {
		return fmt.Errorf("failed to insert thread: %w", err)
	}

	return nil
}

// Readthread retrieves a thread by ID.
func (db *DB) ReadThread(id string) (*Thread, error) {
	query := `SELECT id, download_started, download_completed, work_started, work_completed, solution_proposed, verification_started, solution_revealed, submition_started  FROM threads WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var thread Thread
	if err := row.Scan(&thread.ID, &thread.DownloadStarted, &thread.DownloadCompleted, &thread.WorkStarted, &thread.WorkCompleted, &thread.SolutionProposed, &thread.VerificationStarted, &thread.SolutionRevealed, &thread.SubmitionStarted); err != nil {
		if err == sql.ErrNoRows {
			// thead doesn't exists, so we insert it
			db.AddThread(id)
			return &Thread{ID: id, DownloadStarted: false, DownloadCompleted: false, WorkStarted: false, WorkCompleted: false, SolutionProposed: false, VerificationStarted: false, SubmitionStarted: false}, nil
		}
		return nil, fmt.Errorf("failed to read thread: %w", err)
	}

	return &thread, nil
}

// Updatethread updates a task's information.
func (db *DB) UpdateThread(id string, downloadStarted, downloadCompleted, workStarted, workCompleted, solProposed, verificationStarted, solutionRevealed bool, submitionStarted bool) error {
	updateQuery := `UPDATE threads SET download_started = ?, download_completed = ?, work_started = ?, work_completed = ?, solution_proposed = ?, verification_started = ? , solution_revealed = ?, submition_started = ? WHERE id = ?`
	_, err := db.conn.Exec(updateQuery, downloadStarted, downloadCompleted, workStarted, workCompleted, solProposed, verificationStarted, solutionRevealed, submitionStarted, id)
	if err != nil {
		return fmt.Errorf("failed to update thread: %w", err)
	}
	return nil
}

// Deletethread deletes a thread by ID.
func (db *DB) DeleteThread(id string) error {
	deleteQuery := `DELETE FROM threads WHERE id = ?`
	_, err := db.conn.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete thread: %w", err)
	}
	return nil
}

// Createthread inserts a new thread into the database.
func (db *DB) Addworker(address string) error {
	insertQuery := `INSERT INTO workers (address, registered) VALUES (?, true)`
	_, err := db.conn.Exec(insertQuery, address)
	if err != nil {
		return fmt.Errorf("failed to insert worker: %w", err)
	}

	return nil
}

// Readthread retrieves a thread by ID.
func (db *DB) IsWorkerRegistered(address string) (bool, error) {
	query := `SELECT address, registered  FROM workers WHERE address = ?`
	row := db.conn.QueryRow(query, address)

	var worker Worker
	if err := row.Scan(&worker.Address, &worker.Registered); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to read thread: %w", err)
	}

	return true, nil
}

func (db *DB) DeleteWorker(address string) error {
	deleteQuery := `DELETE FROM workers WHERE address = ?`
	_, err := db.conn.Exec(deleteQuery, address)
	if err != nil {
		return fmt.Errorf("failed to delete worker: %w", err)
	}
	return nil
}

// inserts a new log entry
func (db *DB) AddLogEntry(threadId, log string, timestamp, severity int64) error {
	insertQuery := `INSERT INTO logs (threadId, log, timestamp, severity) VALUES (?,?,?,?)`
	videoUpscalerLogger.Logger.Info("inserting log %s", log)
	_, err := db.conn.Exec(insertQuery, threadId, log, timestamp, severity)
	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}

	return nil
}

func (db *DB) ReadLogs(threadId string) []LogEntry {
	query := `SELECT log, timestamp, severity FROM logs WHERE threadId = ? ORDER BY timestamp`
	rows, _ := db.conn.Query(query, threadId)

	var logs []LogEntry
	for rows.Next() { // Iterate and fetch the records from result cursor
		log := LogEntry{}
		err := rows.Scan(&log.Log, &log.Timestamp, &log.Severity)
		if err != nil {
			videoUpscalerLogger.Logger.Error(err.Error())
		}
		logs = append(logs, log)
	}
	return logs
}

// Createthread inserts a new thread into the database.
func (db *DB) AddIPFSWorker(address string) error {
	insertQuery := `INSERT INTO ipfs (address, added) VALUES (?, true)`
	_, err := db.conn.Exec(insertQuery, address)
	if err != nil {
		return fmt.Errorf("failed to insert worker: %w", err)
	}
	return nil
}

// Readthread retrieves a thread by ID.
func (db *DB) IsIPFSWorkerAdded(address string) (bool, error) {
	query := `SELECT added FROM ipfs WHERE address = ?`
	row := db.conn.QueryRow(query, address)

	var added sql.NullBool
	if err := row.Scan(&added); err != nil {
		if err == sql.ErrNoRows {
			return false, nil // No worker found, returning false
		}
		return false, fmt.Errorf("failed to read IPFS worker status: %w", err)
	}

	// If the value is NULL, treat it as "not added" (false)
	return added.Valid && added.Bool, nil
}

func (db *DB) AddRenderDuration(threadId string, threadNumber, durationInSeconds int) error {
	insertQuery := `INSERT INTO render_times (thread_id, frame_number, render_duration) VALUES (?,?,?)`
	_, err := db.conn.Exec(insertQuery, threadId, threadNumber, durationInSeconds)
	if err != nil {
		videoUpscalerLogger.Logger.Error("failed to insert render duration entry: %s", err.Error())
		return fmt.Errorf("failed to insert render duration entry: %w", err)
	}

	return nil
}

// inserts a new log entry
func (db *DB) GetAverageRenderTime(threadId string) (int, error) {
	query := `SELECT CAST(AVG(render_duration) AS INT)  FROM render_times WHERE thread_id = ?`
	row := db.conn.QueryRow(query, threadId)

	var avg int
	if err := row.Scan(&avg); err != nil {
		videoUpscalerLogger.Logger.Error("failed to read render_times: %s", err.Error())
		return 0, fmt.Errorf("failed to read thread: %w", err)
	}

	return avg, nil
}
