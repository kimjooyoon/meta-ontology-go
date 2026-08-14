package lsp

func (server *Server) lifecycleState() (shutdown, exited bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.shutdown, server.exited
}

func (server *Server) isInitialized() bool {
	server.mu.RLock()
	defer server.mu.RUnlock()
	return server.initialized
}

func (server *Server) markExited() {
	server.mu.Lock()
	server.exited = true
	server.mu.Unlock()
}
