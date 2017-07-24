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
	"ferryd/core"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
	"runtime"
	"sync"
)

// A Processor is responsible for the main dispatch and bulking of jobs
// to ensure they're handled in the most optimal fashion.
type Processor struct {
	manager        *core.Manager
	sequentialjobs chan Runnable
	backgroundJobs chan Runnable
	quit           chan bool
	mut            *sync.Mutex
	wg             *sync.WaitGroup
	closed         bool
	njobs          int
}

// NewProcessor will return a new Processor with the specified number
// of jobs. Note that "njobs" only refers to the number of *background jobs*,
// the majority of operations will run sequentially
func NewProcessor(m *core.Manager, njobs int) *Processor {
	if njobs < 0 {
		njobs = runtime.NumCPU()
	}

	fmt.Fprintf(os.Stderr, "Capped backgroundJobs to %d\n", njobs)

	ret := &Processor{
		manager:        m,
		sequentialjobs: make(chan Runnable),
		backgroundJobs: make(chan Runnable),
		quit:           make(chan bool, 1+njobs),
		mut:            &sync.Mutex{},
		wg:             &sync.WaitGroup{},
		closed:         false,
		njobs:          njobs,
	}
	return ret
}

// Close an existing Processor, waiting for all jobs to complete
func (j *Processor) Close() {
	j.mut.Lock()
	defer j.mut.Unlock()
	if j.closed {
		return
	}
	j.closed = true

	// Disallow further messaging
	close(j.sequentialjobs)
	close(j.backgroundJobs)

	// Ensure all goroutines get the quit broadcast
	for i := 0; i < j.njobs+1; i++ {
		j.quit <- true
	}
	j.wg.Wait()
}

// Begin will start the main job processor in parallel
func (j *Processor) Begin() {
	j.mut.Lock()
	defer j.mut.Unlock()
	if j.closed {
		return
	}
	j.wg.Add(2)
	go j.processSequentialQueue()
	go j.processBackgroundQueue()
}

// reportError will report a failed job to the log
func (j *Processor) reportError(job Runnable, e error) {
	log.WithFields(log.Fields{
		"error": e,
		"type":  reflect.TypeOf(job).Elem().Name(),
	}).Error("Job failed with error")
}

// processSequentialQueue is responsible for dealing with the sequential queue
func (j *Processor) processSequentialQueue() {
	defer j.wg.Done()

	for {
		select {
		case job := <-j.sequentialjobs:
			if job == nil {
				return
			}
			if err := job.Perform(j.manager); err != nil {
				j.reportError(job, err)
			}
		case <-j.quit:
			return
		}
	}
}

// processBackgroundQueue will set up the background workers which will block
// waiting for non-sequential work that cannot run on the main queue, however
// it may put work back on the sequential queue.
func (j *Processor) processBackgroundQueue() {
	defer j.wg.Done()
	j.wg.Add(j.njobs)

	for i := 0; i < j.njobs; i++ {
		go j.backgroundWorker()
	}
}

// backgroundWorker will handle the non sequential tasks as and when they come
// in. The majority of tasks will be sequential on the main queue, so we're free
// to spend more CPU time here dealing with large tasks like the construction
// of delta packages.
func (j *Processor) backgroundWorker() {
	defer j.wg.Done()

	for {
		select {
		case job := <-j.backgroundJobs:
			if job == nil {
				return
			}
			if err := job.Perform(j.manager); err != nil {
				j.reportError(job, err)
			}
		case <-j.quit:
			return
		}
	}
}

// PushJob will take the new job and push it to the appropriate queing system
// For sanity reasons this will lock on the new job add, even if the processing
// is then parallel.
func (j *Processor) PushJob(job Runnable) {
	j.mut.Lock()
	defer j.mut.Unlock()

	if j == nil {
		panic("passed nil job!")
	}

	// TODO: Add descriptions to the Job type and emit to the log
	if job.IsSequential() {
		j.sequentialjobs <- job
	} else {
		j.backgroundJobs <- job
	}
}
