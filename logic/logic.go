package logic

import (
	"fmt"
	"math/rand"
	"practice/domain"
	"sync"
	"time"
)

type taskstore struct {
	tasks  map[int]*domain.Task
	mu     sync.Mutex
	nextid int
	Wg     sync.WaitGroup
}

var Store = taskstore{
	tasks:  make(map[int]*domain.Task),
	nextid: 1,
}

func (s *taskstore) Addtasks(tasks []domain.Task) {
	s.mu.Lock()
	for i := range tasks {
		tasks[i].Id = s.nextid
		tasks[i].Status = domain.Pending
		tasks[i].Createdat = time.Now()
		s.tasks[s.nextid] = &tasks[i]
		s.nextid++
	}
	s.mu.Unlock()
}

func (s *taskstore) Updatetasks(id int, status domain.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, ok := s.tasks[id]; ok {
		task.Status = status
		if status == domain.Done {
			task.Completedat = time.Now()
		}
	}
}

func (s *taskstore) Gettask(id int) (*domain.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	return task, ok
}

func workerpool(jobqueue <-chan *domain.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range jobqueue {
		Store.Updatetasks(task.Id, domain.Processing)
		duration := time.Duration(rand.Intn(5)+1) * time.Second
		fmt.Printf("processing task id: %d and it will take %v\n", task.Id, duration)
		time.Sleep(duration)
		Store.Updatetasks(task.Id, domain.Done)
		fmt.Printf("the task id:%d is completed\n", task.Id)
	}
}

func Processtasks(tasks []domain.Task) {
	jobqueue := make(chan *domain.Task, len(tasks))

	Store.Wg.Add(1)
	defer Store.Wg.Done()

	var localwg sync.WaitGroup
	for i := 0; i < 3; i++ {
		localwg.Add(1)
		go workerpool(jobqueue, &localwg)
	}

	Store.mu.Lock()
	for _, task := range Store.tasks {
		jobqueue <- task
	}
	Store.mu.Unlock()

	close(jobqueue)
	localwg.Wait()
}
