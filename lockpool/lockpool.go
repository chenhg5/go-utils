package lockpool

import "sync"

type LockPool struct {
	List     map[int]*sync.Mutex
	LockPool sync.Pool
}

var Pool *LockPool

var once sync.Once

var LockOfPool sync.Mutex

func GetPacketLockList() *LockPool {
	once.Do(func() {
		Pool = &LockPool{
			make(map[int]*sync.Mutex, 1000), // 取一个多大的值合适

			sync.Pool{
				New: func() interface{} {
					return new(sync.Mutex)
				},
			},
		}
	})

	return Pool
}

func (lock *LockPool) GetLockById(id int) *sync.Mutex {
	LockOfPool.Lock()
	defer LockOfPool.Unlock()
	return (*lock).List[id]
}

func (lock *LockPool) SetALock(id int) {
	if !lock.CheckLock(id) {
		LockOfPool.Lock()
		mu := (*lock).LockPool.Get().(*sync.Mutex)
		(*lock).List[id] = mu
		LockOfPool.Unlock()
	}
}

func (lock *LockPool) ResetALock(id int) {
	if !lock.CheckLock(id) {
		LockOfPool.Lock()
		mu := (*lock).LockPool.Get().(*sync.Mutex)
		(*lock).List[id] = mu
		LockOfPool.Unlock()
	} else {
		lock.DeleteLock(id)
		LockOfPool.Lock()
		mu := (*lock).LockPool.Get().(*sync.Mutex)
		(*lock).List[id] = mu
		LockOfPool.Unlock()
	}
}

func (lock *LockPool) LockById(id int) {
	(*((*lock).List[id])).Lock()
}

func (lock *LockPool) UnLockById(id int) {
	(*((*lock).List[id])).Unlock()
}

func (lock *LockPool) DeleteLock(id int) {
	LockOfPool.Lock()
	defer LockOfPool.Unlock()
	(*lock).LockPool.Put((*lock).List[id])
	(*lock).List[id] = nil
}

func (lock *LockPool) CheckLock(id int) bool {
	LockOfPool.Lock()
	defer LockOfPool.Unlock()
	if val, ok := (*lock).List[id]; !ok {
		return false
	} else if val == nil {
		return false
	} else {
		return true
	}
}

