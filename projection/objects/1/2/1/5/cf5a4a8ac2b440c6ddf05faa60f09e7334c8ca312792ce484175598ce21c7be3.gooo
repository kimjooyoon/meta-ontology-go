package cache

func (c *Cache) acquireKey(key Key) (func(), error) {
	if err := c.validatePathKey(key); err != nil {
		return nil, err
	}
	name := key.String()
	c.locksMu.Lock()
	lock := c.locks[name]
	if lock == nil {
		lock = &entryLock{}
		c.locks[name] = lock
	}
	lock.refs++
	c.locksMu.Unlock()
	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		c.locksMu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(c.locks, name)
		}
		c.locksMu.Unlock()
	}, nil
}
