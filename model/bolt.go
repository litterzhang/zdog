package model

import (
	"fmt"
	"time"
	"encoding/binary"
	"strings"
	"bytes"
	"encoding/json"
	"github.com/labstack/gommon/log"
	bolt "go.etcd.io/bbolt"
)

const (
	MODULE_NAME  = "zdog_db"
	DB_PATH      = "/home/zhang/data/zdog/bolt.db"
	TIMEOUT      = 1 * time.Second
	INDEX_BUCKET = "_index"
	INDEX_SPLIT  = "#"
)

type BoltDb struct {
	db     *bolt.DB
}

type Index struct {
	bucket string
	indexs []string
}

func (i *Index) key() string {
	return strings.Join(i.indexs, INDEX_SPLIT)
}

// global db
var (
	db *BoltDb
	logger *log.Logger
	indexs map[string][]*Index
)

func BindIndex(bname string, keys ...string) {
	if len(keys) < 1 {
		return
	}
	index := &Index {
		bucket: bname,
		indexs: keys,
	}
	bindexs, ok := indexs[bname]
	if !ok {
		
		bindexs = []*Index {index, }
	} else {
		index_key := index.key()
		for _, _index := range bindexs {
			if index_key == _index.key() {
				return 
			}
		}
		bindexs = append(bindexs, index)
	}
	indexs[bname] = bindexs

	// dump to db
	saveIndexs(bindexs)
}

func open() *bolt.DB {
	if db != nil {
		return db.db
	}
	logger = log.New(MODULE_NAME) 
	opts := &bolt.Options{
		Timeout: TIMEOUT,
	}
	boltdb, err := bolt.Open(DB_PATH, 0600, opts) 
	if err != nil {
		logger.Fatal("open bolt db error", err)
		db = nil
	} else {
		db = &BoltDb{
			db:     boltdb,
		}
	}
	return db.db
}

// save index
func saveIndexs(bindexs []*Index) {
	db := open()
	err := db.Update(func(tx *bolt.Tx) error {
		__indexs := make([]string, 0)
		for _, index := range bindexs {
			v := strings.Join(index.indexs, INDEX_SPLIT)
			__indexs = append(__indexs, v)
		}
		b, err := tx.CreateBucketIfNotExists([]byte(INDEX_BUCKET))
		if err != nil {
			return err
		}
		buf, err := json.Marshal(__indexs)
		err = b.Put([]byte(bindexs[0].bucket), buf)
		return err
	})
	if err != nil {
		logger.Errorf("save indexs[%v] error -> %v", bindexs, err)
	}
}

// load indexs from database bucket indexs
func loadIndexs() {
	db := open()
	indexs = make(map[string][]*Index)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(INDEX_BUCKET))
		if b == nil {
			logger.Infof("bucket[%s] not exist", INDEX_BUCKET)
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			key := string(k)
			var _indexs []string
			err := json.Unmarshal(v, &_indexs)
			if err != nil {
				logger.Errorf("parse index[%s] error -> %v", k, err)
				continue
			}
			__indexs := make([]*Index, len(_indexs)) 
			// parse index
			for i, _index := range _indexs {
				sps :=strings.Split(_index, INDEX_SPLIT) 
				index := &Index {
					bucket: key,
					indexs: sps,
				}
				__indexs[i] = index
			}
			indexs[key] = __indexs
		}
		return nil
	})
}

func CloseDb() {
	db.db.Close()
}

func OpenDb() {
	loadIndexs()
}

// insert 
func Insert(bucket string, kvs map[string]interface{}) {
	db := open()
	var id int
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			logger.Error("open bucket error", bucket, err)
			return err
		}
		buf, err := json.Marshal(kvs)
		if err != nil {
			logger.Error("json marshal error ", kvs, err)
			return err
		}
		// query if records in db
		ids := queryIds(bucket, kvs)
		if len(ids) == 0 {
			_id, _ := b.NextSequence()
			id = int(_id)
			kvs["__id__"] = id
			err = b.Put(itob(id), buf)
		} else {
			for _, _id := range ids {
				b.Put(_id, buf)
			}
			err = fmt.Errorf("updated records in db")
		}
		return err
	})

	if err != nil {
		logger.Errorf("insert record[%v] to db[%s] error -> %v", kvs, bucket, err)
		return
	}
	bindexs, ok := indexs[bucket]
	ibucket := fmt.Sprintf("%s-%s", INDEX_BUCKET, bucket)
	// put indexs
	if ok {
		db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(ibucket))
			if err != nil {
				logger.Error("open bucket error", ibucket, err)
				return err
			}
			for _, index := range bindexs {
				f := index.indexs[0]
				if _, ok := kvs[f]; !ok {
					 continue
				}
				_index := "" 
				for _, k := range index.indexs {
					v, _ := kvs[k]
					_index += fmt.Sprintf("%s=%v", k, v)
				}
				_index += string(id)
				err = b.Put([]byte(_index), itob(id))
				if err != nil {
					logger.Error("put error ", ibucket, _index, err)
					continue
				}
			}
			return nil
		})
	}
}

func queryIds(bucket string, kvs map[string]interface{}) [][]byte {
	db := open()
	// try find by index
	bindexs, ok := indexs[bucket]
	var ids [][]byte = [][]byte {}
	if ok {
		ibucket := fmt.Sprintf("%s-%s", INDEX_BUCKET, bucket)	
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(ibucket))
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for _, index := range bindexs {
				f := index.indexs[0]
				if _, ok := kvs[f]; !ok {
					 continue
				}
				_index := []byte(fmt.Sprintf("%s=%v", f, kvs[f])) 
				for k, v := c.Seek(_index); k != nil && bytes.HasPrefix(k, _index); k, v = c.Next() {
					ids = append(ids, v)
				}
			}
			return nil
		})
	}
	return ids
}

// query
func Query(bucket string, kvs map[string]interface{}) []map[string]interface{} {
	db := open()
	ids := queryIds(bucket, kvs)
	vs := make([]map[string]interface{}, 0)
	// query data by id
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		for _, bid := range ids {
			v := b.Get(bid)

			var vv map[string]interface{}
			err := json.Unmarshal(v, &vv)
			if err != nil {
				logger.Errorf("unmarshal data[%s] error -> %v'", v, err)
				continue
			}
			vs = append(vs, vv)
		}
		return nil
	})
	return vs
}

func Get(bucket string, kvs map[string]interface{}) map[string]interface{} {
	vs := Query(bucket, kvs)
	if len(vs) == 0 {
		return nil
	}
	for _, v := range vs {
		for k, _v := range kvs {
			if __v, ok := v[k]; !ok || __v != _v {
				continue
			}
			return v
		}
	}
	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(v))
    return b
}


// itob returns an 8-byte big endian representation of v.
func btoi(b []byte) int {
	v := int(binary.BigEndian.Uint64(b))
    return v
}
