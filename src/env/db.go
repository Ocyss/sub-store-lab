package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

type dbLogger struct{}

var _ badger.Logger = (*dbLogger)(nil)

func (dbLogger) Errorf(msg string, args ...any) {
	slog.Error(fmt.Sprintf(msg, args...))
}

func (dbLogger) Warningf(msg string, args ...any) {
	slog.Warn(fmt.Sprintf(msg, args...))
}

func (dbLogger) Infof(msg string, args ...any) {
	slog.Info(fmt.Sprintf(msg, args...))
}

func (dbLogger) Debugf(msg string, args ...any) {
	slog.Debug(fmt.Sprintf(msg, args...))
}

func initDB() {
	var err error
	db, err = badger.Open(
		badger.
			DefaultOptions(filepath.Join(Conf.DataDir, "sub-store-lab.badger")).
			WithLogger(dbLogger{}),
	)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		panic(err)
	}
}

func GetDB() *badger.DB {
	return db
}

func CloseDB() error {
	if db != nil {
		err := db.Close()
		db = nil
		return err
	}
	return nil
}

type DbPrefixFn[T any] = func(txn *badger.Txn, k []byte, v T) error

func dbPrefixFn[T any](fn DbPrefixFn[T], prefix []byte, all bool) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		var errs error
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var data T
			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &data)
			})
			if err != nil {
				if all {
					return fmt.Errorf("get db prefix[%s] key:%s err: %w", string(prefix), string(k), errs)
				} else {
					errs = errors.Join(errs, fmt.Errorf("err[%s]: %w", string(k), err))
				}
			}
			err = fn(txn, k, data)
			if err != nil {
				if all {
					return fmt.Errorf("get db prefix[%s] key:%s handleErr: %w", string(prefix), string(k), errs)
				} else {
					errs = errors.Join(errs, fmt.Errorf("handleErr[%s]: %w", string(k), err))
				}
			}
		}
		if errs != nil {
			return fmt.Errorf("get db prefix[%s] errs: %w", string(prefix), errs)
		}
		return nil
	}
}

func QueryDb[T any](key []byte) (T, error) {
	var result T
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &result); err != nil {
				return err
			}
			return nil
		})
	})
	return result, err
}

func QueryDbPrefix[T any](fn DbPrefixFn[T], prefix []byte, all bool) error {
	err := db.View(dbPrefixFn(fn, prefix, all))
	return err
}

func UpdateDbPrefix[T any](fn DbPrefixFn[T], prefix []byte, all bool) error {
	err := db.Update(dbPrefixFn(fn, prefix, all))
	return err
}
