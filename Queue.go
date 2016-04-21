package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Queue struct {
	state   map[int]*Node
	pushing chan *push
	popping chan *pop
}

type Node struct {
	data interface{}
	next *Node
}

type push struct {
	key  int
	val  *Node
	resp chan bool
}

type pop struct {
	key  int
	resp chan *Node
}

func (q *Queue) Count() int {
	if q == nil {
		return 0
	}

	return len(q.state)
}

// TODO: Channelify
func (q *Queue) Push(item interface{}) <-chan bool {

	// create push item to add to channel
	push := &push{
		key: q.Count(),
		val: &Node{
			data: item,
		},
		resp: make(chan bool),
	}

	q.pushing <- push

	// wait
	return push.resp
}

// Pop returns a channel that you can select IO on
func (q *Queue) Pop() <-chan *Node {

	// create pop item
	pop := &pop{
		key:  0,
		resp: make(chan *Node),
	}

	// popping item
	q.popping <- pop

	return pop.resp
}

// Construct creates a new queue which is goroutine-safe
func Construct() *Queue {

	// Create queue with push and pop channels
	q := &Queue{}
	q.pushing = make(chan *push)
	q.popping = make(chan *pop)

	// goroutine that handles state
	go func() {
		if q.state == nil {
			q.state = make(map[int]*Node)
		}

		for {

			select {

			// someone called pop, wants item
			case pop := <-q.popping:
				// set response as *node value
				pop.resp <- q.state[pop.key]

				// remove item
				delete(q.state, pop.key)

				// someone called push, wants to add new item
			case push := <-q.pushing:
				// add new *node value by key
				q.state[push.key] = push.val

				// success
				push.resp <- true
			}
		}

	}()

	return q
}

func main() {
	queue := Construct()

	// start reading and writing
	go write(queue)
	go read(queue)

	select {}

}

func read(queue *Queue) {

	// start ze popping
	for pop := range queue.Pop() {
		fmt.Printf("Popping val: %s on time %s\n", pop.data, time.Now())

		time.Sleep(time.Second * 1)
	}
	fmt.Printf("read done?")
}

func write(queue *Queue) {

	for index := 0; index < 3; index++ {

		val := rand.Intn(100)
		if <-queue.Push(val) == true {
			fmt.Printf("Value: %d pushed, count is: %d, on time %s\n", val, queue.Count(), time.Now())
		} else {
			fmt.Printf("Could not push item!\n")
		}

		time.Sleep(time.Second * 3)
	}
}
