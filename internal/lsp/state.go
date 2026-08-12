package lsp

func (server *Server) lifecycleState() (shutdown, exited bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.shutdown, server.exited
}

func (server *Server) markExited() {
	server.mu.Lock()
	server.exited = true
	server.mu.Unlock()
}
