//
// Copyright © 2017 Ikey Doherty <ikey@solus-project.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package jobs

import (
	"encoding/binary"
	"errors"
	"github.com/boltdb/bolt"
)

var (
	// BucketAsyncJobs holds all asynchronous jobs
	BucketAsyncJobs = []byte("Async")

	// BucketSequentialJobs holds all sequential jobs
	BucketSequentialJobs = []byte("Sync")

	// BucketRootJobs is the parent job bucket
	BucketRootJobs = []byte("JobRoot")

	// ErrEmptyQueue is returned to indicate a job is not available yet
	ErrEmptyQueue = errors.New("Queue is empty")
)

// JobStore handles the storage and manipulation of incomplete jobs
type JobStore struct {
	db *bolt.DB
}

// NewStore creates a fully initialized JobStore and sets up Bolt Buckets as needed
func NewStore(db *bolt.DB) (s *JobStore, err error) {
	s = &JobStore{db}
	err = s.setup()
	return
}

// Setup makes sure that all the necessary buckets exist and have valid contents
func (s *JobStore) setup() error {
	buckets := [][]byte{
		BucketAsyncJobs,
		BucketSequentialJobs,
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		rootBucket, err := tx.CreateBucketIfNotExists(BucketRootJobs)
		if err != nil {
			return err
		}
		for _, b := range buckets {
			if _, err := rootBucket.CreateBucketIfNotExists(b); err != nil {
				return err
			}
			return nil
		}
		return nil
	})
}

// ClaimAsyncJob gets the first available asynchronous job, if one exists
func (s *JobStore) ClaimAsyncJob() (*JobEntry, error) {
	var job *JobEntry

	err := s.db.Update(func(tx *bolt.Tx) error {
		async := tx.Bucket(BucketRootJobs).Bucket(BucketAsyncJobs)
		cursor := async.Cursor()
		id, value := cursor.First()
		var newJ []byte
		for id != nil {
			j, err := Deserialize(value)
			if err != nil {
				return err
			}
			if !j.Claimed {
				j.Claimed = true

				// Serialise the new guy
				newJ, err = j.Serialize()
				if err != nil {
					return err
				}

				// Put the new guy back in
				if err = async.Put(id, newJ); err != nil {
					return err
				}

				// Got a usable job now.
				job = j
				return nil
			}
			id, value = cursor.Next()
		}
		// No available jobs to peek
		return ErrEmptyQueue
	})

	if err != nil {
		return nil, err
	}

	return job, nil
}

// ClaimSequentialJob gets the first available synchronous job, if one exists
func (s *JobStore) ClaimSequentialJob() (*JobEntry, error) {
	var job *JobEntry

	err := s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(BucketRootJobs).Bucket(BucketSequentialJobs).Cursor()
		id, value := cursor.First()
		if id == nil {
			return ErrEmptyQueue
		}
		j, e := Deserialize(value)
		if e != nil {
			return e
		}
		// Store private ID field
		j.id = id
		job = j
		return nil
	})

	if err != nil {
		return nil, err
	}

	return job, nil
}

// RetireAsyncJob removes a completed asynchronous job
func (s *JobStore) RetireAsyncJob(j *JobEntry) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BucketRootJobs).Bucket(BucketAsyncJobs).Delete(j.id)
	})
}

// RetireSequentialJob removes a completed synchronous job
func (s *JobStore) RetireSequentialJob(j *JobEntry) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(BucketRootJobs).Bucket(BucketSequentialJobs).Delete(j.id)
	})
}

// pushJobInternal is identical between sync and async jobs, it
// just needs to know which bucket to store the job in.
func (s *JobStore) pushJobInternal(j *JobEntry, bk []byte) error {
	j.Claimed = false

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketRootJobs).Bucket(bk)
		id, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		// Adapted from boltdb itob example code
		j.id = make([]byte, 8)
		binary.BigEndian.PutUint64(j.id, uint64(id))
		blob, err := j.Serialize()
		if err != nil {
			return err
		}
		return bucket.Put(j.id, blob)
	})
}

// PushSequentialJob will enqueue a new sequential job
func (s *JobStore) PushSequentialJob(j *JobEntry) error {
	return s.pushJobInternal(j, BucketSequentialJobs)
}

// PushAsyncJob will enqueue a new asynchronous job
func (s *JobStore) PushAsyncJob(j *JobEntry) error {
	return s.pushJobInternal(j, BucketAsyncJobs)
}