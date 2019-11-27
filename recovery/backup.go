package recovery

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/screepers/screeps-launcher/v1/launcher"
)

type backup struct {
	Redis redisData
	Mongo mongoData
	Files fileData
}

// Recovery handles backups
type Recovery struct {
	config *launcher.Config
}

// UsesScreepsmodMongo checks if screepsmod-mongo is installed
func (r *Recovery) UsesScreepsmodMongo() bool {
	for _, mod := range r.config.Mods {
		if mod == "screepsmod-mongo" {
			return true
		}
	}
	return false
}

// New creates a new Recovery instance
func New() *Recovery {
	return &Recovery{}
}

// Backup performas a backup of the server
func (r *Recovery) Backup() error {
	if len(os.Args) < 3 {
		fmt.Print("Usage: screeps-launcher backup <file>")
		os.Exit(1)
	}
	file := os.Args[2]
	err := r.BackupFile(file)
	if err != nil {
		return err
	}
	log.Printf("Backup complete")
	return nil
}

// Restore restores a backup of the server
func (r *Recovery) Restore() error {
	if len(os.Args) < 3 {
		fmt.Print("Usage: screeps-launcher restore <file>")
		os.Exit(1)
	}
	file := os.Args[2]
	err := r.RestoreFile(file)
	if err != nil {
		return err
	}
	log.Printf("Restore complete")
	return nil
}

// BackupFile saves a backup to a file
func (r *Recovery) BackupFile(filename string) error {
	b := backup{}
	c := launcher.NewConfig()
	c.GetConfig("")
	r.config = c
	var err error
	if r.UsesScreepsmodMongo() {
		b.Redis, err = r.redisBackup()
		if err != nil {
			return err
		}
		b.Mongo, err = r.mongoBackup()
		if err != nil {
			return err
		}
	}
	b.Files, err = r.filesBackup()
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	gz := gzip.NewWriter(file)
	defer gz.Close()
	enc := gob.NewEncoder(gz)
	enc.Encode(b)
	return nil
}

// RestoreFile restores a backup from a file
func (r *Recovery) RestoreFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	var b backup
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	dec := gob.NewDecoder(gz)
	dec.Decode(&b)
	if err := r.filesRestore(b.Files); err != nil {
		return err
	}
	c := launcher.NewConfig()
	c.GetConfig("")
	r.config = c
	if r.UsesScreepsmodMongo() {
		if err := r.redisRestore(b.Redis); err != nil {
			return err
		}
		if err := r.mongoRestore(b.Mongo); err != nil {
			return err
		}
	}
	return nil
}
