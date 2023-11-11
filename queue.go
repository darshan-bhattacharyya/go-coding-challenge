package main

type Queue struct {
	Queue [60]int64 `json:"queue"`
}

func (q *Queue) Push(item int64) {
	q.Pop()
	q.Queue[0] = item
}

func (q *Queue) Pop() int64 {
	itemToPop := q.Queue[59]
	for i := 0; i < 59; i++ {
		q.Queue[i+1] = q.Queue[i]
	}
	q.Queue[0] = 0
	return itemToPop
}
