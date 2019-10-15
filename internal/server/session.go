package server

//Session in server side
type Session struct {
	sessionID uint32
	server    *Server
}

//Handler of server commands
type Handler struct {
}
