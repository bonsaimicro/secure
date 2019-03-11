package database

import (
	"bytes"
	"io"
	"log"
	"os"
	"secure/logger"

	"github.com/dgraph-io/badger"
	"github.com/minio/sio"
	"golang.org/x/crypto/argon2"
)

// The Datastore interface
type Datastore interface {
	Close() error
	AddUser(*User) (*User, error)
	FindUser(email, password string) (*User, error)
}

func (d *datastore) Close() error {
	return d.db.Close()
}

func (d *datastore) Add(bucket string, id string, obj Modeler) (*Modeler, error) {
	go d.l.LogDBRequest("INSERT INTO "+bucket, id)
	k := []byte(bucket + ":" + id)

	err := d.db.Update(func(txn *badger.Txn) error {
		v, err := obj.encode()
		if err != nil {
			return err
		}

		encrypted, err := d.encrypt(v)
		if err != nil {
			return err
		}

		err = txn.Set([]byte(k), encrypted)

		return err
	})

	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (d *datastore) Fetch(bucket, id string) ([]byte, error) {
	k := bucket + ":" + id
	go d.l.LogDBRequest(bucket, k)
	var valCopy []byte

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))

		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})

		return err
	})

	if err != nil {
		return []byte{}, err
	}
	decryptedValue, err := d.decrypt(bytes.NewReader(valCopy))
	if err != nil {
		return []byte{}, err
	}

	return decryptedValue, nil
}

func (d *datastore) Update(bucket, aID, id string, obj Modeler) error {
	k := bucket + ":" + aID + ":" + id
	go d.l.LogDBRequest("UPDATE "+bucket, k)
	err := d.db.Update(func(txn *badger.Txn) error {
		v, err := obj.encode()
		if err != nil {
			return err
		}

		encrypted, err := d.encrypt(v)
		if err != nil {
			return err
		}

		err = txn.Set([]byte(k), encrypted)

		return err
	})

	return err
}

type datastore struct {
	db    *badger.DB
	l     logger.Logger
	idKey []byte
}

// New sets up the db connection
func New(path string, l logger.Logger) (Datastore, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)

	if err != nil {
		log.Fatal(err)
	}

	pw := os.Getenv("DARE_PASSWORD")
	salt := os.Getenv("DARE_SALT")
	if pw == "" || salt == "" {
		log.Fatal("DARE_PASSWORD & DARE_SALT env vars not set")
	}
	return &datastore{db, l, argon2.IDKey([]byte(pw), []byte(salt), 1, 64*1024, 4, 32)}, nil
}

func (d *datastore) encrypt(v io.Reader) ([]byte, error) {
	encrypted, err := sio.EncryptReader(v, sio.Config{Key: d.idKey})
	if err != nil {
		return []byte(""), err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(encrypted)
	return buf.Bytes(), nil
}

func (d *datastore) decrypt(v io.Reader) ([]byte, error) {
	encrypted, err := sio.DecryptReader(v, sio.Config{Key: d.idKey})
	if err != nil {
		return []byte(""), err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(encrypted)
	return buf.Bytes(), nil
}
